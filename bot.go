package main

import (
	"fmt"
	jenkins "github.com/bndr/gojenkins"
	"github.com/go-xorm/xorm"
	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-gitee-framework"
	"github.com/opensourceways/community-robot-lib/utils"
	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"time"
)

const botName = "invoke-bot"

type iClient interface{}

func newRobot(cli iClient, jenkinsCli *jenkins.Jenkins, engine *xorm.Engine) *robot {
	return &robot{cli: cli, jenkinsCli: jenkinsCli, engine: engine}
}

type robot struct {
	cli        iClient
	jenkinsCli *jenkins.Jenkins
	engine     *xorm.Engine
}

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) getConfig(cfg config.Config, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(p framework.HandlerRegister) {
	p.RegisterPullRequestHandler(bot.handlePREvent)
	p.RegisterNoteEventHandler(bot.handleNoteEvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, pc config.Config, log *logrus.Entry) error {
	//if sdk.GetPullRequestAction(e) != sdk.ActionOpen && sdk.GetPullRequestAction(e) != sdk.PRActionChangedSourceBranch {
	if sdk.GetPullRequestAction(e) != sdk.ActionOpen {
		if e.GetAction() == "update" && e.GetActionDesc() != sdk.PRActionChangedTargetBranch &&
			e.GetActionDesc() != sdk.PRActionChangedSourceBranch {
			_ = bot.dealNotInvokeWebhooks(log)
		}
		return nil
	}

	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()
	committer := e.GetPRAuthor()
	targetBranch := e.GetPullRequest().GetBase().GetRef()
	sourceBranch := e.GetSourceBranch()
	sourceOrg := e.GetSourceRepo().Project.Namespace
	prCreateAt := e.GetPullRequest().GetCreatedAt()
	rand.Seed(time.Now().Unix())
	randNumber := strconv.Itoa(rand.Int())
	jobName := fmt.Sprintf("multiarch/job/%s/job/trigger/job/%s", org, repo)
	params := map[string]string{
		"giteeRepoName":        repo,
		"giteePullRequestIid":  strconv.Itoa(int(number)),
		"giteeSourceBranch":    sourceBranch,
		"giteeTargetBranch":    targetBranch,
		"giteeSourceNamespace": sourceOrg,
		"giteeTargetNamespace": org,
		"giteeCommitter":       committer,
		"prCreateTime":         prCreateAt,
		"PULL_NUMBER":          strconv.Itoa(int(number)),
		"REPO_OWNER":           org,
		"REPO_NAME":            repo,
		"randSeed":             randNumber,
	}

	merr := utils.NewMultiErrors()

	retry := 0
	for {
		buildNumber, err := bot.jenkinsCli.BuildJob(Ctx, jobName, params)
		if err != nil || buildNumber == 0 {
			retry += 1
		}

		if err == nil && buildNumber != 0 {
			log.Infof("job %s which is triggered by %s has been invoked successfully", jobName,
				fmt.Sprintf("%s/%s/pulls/%d", org, repo, number))
			break
		}

		if retry > 2 {
			log.Errorf("invoke jenkins job %s triggered by %s failed, "+
				"err: %s, retry 3 times, but still failed, sent it to the message queue",
				jobName, fmt.Sprintf("%s/%s/pulls/%d", org, repo, number), err.Error())

			var d = Webhooks{
				Repo:         repo,
				TargetOrg:    org,
				SourceOrg:    sourceOrg,
				TargetBranch: targetBranch,
				SourceBranch: sourceBranch,
				Number:       strconv.Itoa(int(number)),
				Comment:      "",
				CommentId:    "",
				RandNumber:   randNumber,
				TriggerLink:  "",
				TriggerTime:  "",
				Scheduled:    false,
				Deleted:      false,
				PrCreateTime: prCreateAt,
				Committer:    committer,
			}
			err = bot.insertToDB(&d)
			if err != nil {
				log.Errorf("insert details of webhook failed, err: %s", err.Error())
			}

			break
		}
	}

	return merr.Err()
}

func (bot *robot) handleNoteEvent(e *sdk.NoteEvent, pc config.Config, log *logrus.Entry) error {
	if !e.IsPullRequest() ||
		!e.IsPROpen() ||
		!e.IsCreatingCommentEvent() ||
		e.GetComment().GetBody() != "/retest" {
		return nil
	}

	_ = bot.dealNotInvokeWebhooks(log)

	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()
	author := e.GetPRAuthor()
	targetBranch := e.GetPullRequest().GetBase().GetRef()
	sourceBranch := e.GetPullRequest().GetHead().GetRef()
	sourceOrg := e.GetPullRequest().GetHead().Repo.Namespace
	prCreateAt := e.GetPullRequest().GetCreatedAt()
	comment, commentID := e.GetComment().GetBody(), e.GetComment().GetID()
	jobTriggerTime, triggerLink := e.GetComment().GetUpdatedAt(), e.GetComment().GetHtmlUrl()

	jobName := fmt.Sprintf("multiarch/job/%s/job/trigger/job/%s", org, repo)
	rand.Seed(time.Now().Unix())
	randNumber := strconv.Itoa(rand.Int())
	params := map[string]string{
		"giteeRepoName":        repo,
		"giteePullRequestIid":  strconv.Itoa(int(number)),
		"giteeSourceBranch":    sourceBranch,
		"giteeTargetBranch":    targetBranch,
		"giteeSourceNamespace": sourceOrg,
		"giteeTargetNamespace": org,
		"giteeCommitter":       author,
		"prCreateTime":         prCreateAt,
		"PULL_NUMBER":          strconv.Itoa(int(number)),
		"REPO_OWNER":           org,
		"REPO_NAME":            repo,
		"comment":              comment,
		"commentID":            strconv.Itoa(int(commentID)),
		"jobTriggerTime":       jobTriggerTime,
		"triggerLink":          triggerLink,
		"randSeed":             randNumber,
	}

	merr := utils.NewMultiErrors()

	retry := 0
	for {
		buildNumber, err := bot.jenkinsCli.BuildJob(Ctx, jobName, params)
		if err != nil || buildNumber == 0 {
			retry += 1
		}

		if err == nil && buildNumber != 0 {
			log.Infof("job %s which is triggered by %s has been invoked successfully", jobName,
				fmt.Sprintf("%s/%s/pulls/%d/#note_%d", org, repo, number, commentID))
			break
		}

		if retry > 2 {
			log.Errorf("invoke jenkins job %s triggered by %s failed, err: %s, retry 3 times, but still failed", jobName,
				fmt.Sprintf("%s/%s/pulls/%d", org, repo, number), err.Error())

			var d = Webhooks{
				Repo:         repo,
				TargetOrg:    org,
				SourceOrg:    sourceOrg,
				TargetBranch: targetBranch,
				SourceBranch: sourceBranch,
				Number:       strconv.Itoa(int(number)),
				Comment:      comment,
				CommentId:    strconv.Itoa(int(commentID)),
				RandNumber:   randNumber,
				TriggerLink:  triggerLink,
				TriggerTime:  jobTriggerTime,
				Scheduled:    false,
				Deleted:      false,
				PrCreateTime: prCreateAt,
				Committer:    author,
			}
			err = bot.insertToDB(&d)
			if err != nil {
				log.Errorf("insert details of webhook failed, err: %s", err.Error())
			}

			break
		}
	}

	return merr.Err()
}
