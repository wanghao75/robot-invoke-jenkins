package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

func (bot *robot) insertToDB(data *Webhooks) error {
	_, err := bot.engine.Insert(data)
	if err != nil {
		return err
	}

	return nil
}

func (bot *robot) selectNotScheduledFromDB() ([]Webhooks, error) {
	var data []Webhooks

	err := bot.engine.Where("scheduled = ? and deleted = ?", false, false).Find(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (bot *robot) deleteSingleFromDB(id int) error {
	var data = Webhooks{
		Id:      id,
		Deleted: true,
	}

	rows, err := bot.engine.Delete(&data)
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("failed to delete")
	}

	return nil
}

func (bot *robot) softDeleteByTag(id int) error {
	var d = Webhooks{
		Id:        id,
		Deleted:   true,
		Scheduled: true,
	}
	_, err := bot.engine.Update(&d)
	if err != nil {
		return err
	}

	return nil
}

func (bot *robot) dealNotInvokeWebhooks(log *logrus.Entry) error {
	needDealWebhooks, err := bot.selectNotScheduledFromDB()
	if err != nil {
		return err
	}

	if len(needDealWebhooks) == 0 {
		log.Info("no webhooks in database, so skip this step")
		return nil
	}

	for _, w := range needDealWebhooks {
		params := map[string]string{
			"giteeRepoName":        w.Repo,
			"giteePullRequestIid":  w.Number,
			"giteeSourceBranch":    w.SourceBranch,
			"giteeTargetBranch":    w.TargetBranch,
			"giteeSourceNamespace": w.SourceOrg,
			"giteeTargetNamespace": w.TargetOrg,
			"giteeCommitter":       w.Committer,
			"prCreateTime":         w.PrCreateTime,
			"PULL_NUMBER":          w.Number,
			"REPO_OWNER":           w.TargetOrg,
			"REPO_NAME":            w.Repo,
			"comment":              w.Comment,
			"commentID":            w.CommentId,
			"jobTriggerTime":       w.TriggerTime,
			"triggerLink":          w.TriggerLink,
			"randSeed":             w.RandNumber,
		}

		webhookTriggerBy := ""
		if w.Comment != "" {
			webhookTriggerBy = fmt.Sprintf("%s/%s/pulls/%s#note_%s", w.TargetOrg, w.Repo, w.Number, w.CommentId)
		} else {
			webhookTriggerBy = fmt.Sprintf("%s/%s/pulls/%s", w.TargetOrg, w.Repo, w.Number)
		}

		jobName := fmt.Sprintf("multiarch/job/%s/job/trigger/job/%s", w.TargetOrg, w.Repo)
		_, err := bot.jenkinsCli.BuildJob(Ctx, jobName, params)

		if err != nil {
			errLog := fmt.Sprintf("reinvoke jenkins job %s triggered by %s failed, err: %s", jobName,
				webhookTriggerBy, err.Error())
			log.Errorf(errLog)
			continue
		} else {
			log.Infof("reinvoke jenkins job %s triggered by %s success", jobName, webhookTriggerBy)
			err = bot.deleteSingleFromDB(w.Id)
			if err != nil {
				log.Errorf("delete webhook from %s in database failed, err: %s", webhookTriggerBy, err.Error())
			}
		}
	}

	return nil
}
