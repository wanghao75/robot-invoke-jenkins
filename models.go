package main

type Webhooks struct {
	Id           int    `xorm:"id serial pk not null" json:"id"`
	Repo         string `xorm:"'repo' varchar(50) not null" json:"repo"`
	TargetOrg    string `xorm:"'target_org' varchar(50) not null" json:"target_org"`
	TargetBranch string `xorm:"'target_branch' varchar(50) not null" json:"target_branch"`
	SourceOrg    string `xorm:"'source_org' varchar(50) not null" json:"source_org"`
	SourceBranch string `xorm:"'source_branch' varchar(50) not null" json:"source_branch"`
	Number       string `xorm:"'number' varchar(50) not null" json:"number"`
	Comment      string `xorm:"'comment' varchar(20)" json:"comment"`
	CommentId    string `xorm:"'comment_id' varchar(50)" json:"comment_id"`
	TriggerTime  string `xorm:"'trigger_time' varchar(128)" json:"trigger_time"`
	TriggerLink  string `xorm:"'trigger_link' varchar(128)" json:"trigger_link"`
	RandNumber   string `xorm:"'rand_number' varchar(50) not null" json:"rand_number"`
	PrCreateTime string `xorm:"'pr_create_time' varchar(50) not null" json:"pr_create_time"`
	Committer    string `xorm:"'committer' varchar(128) not null" json:"committer"`
	Scheduled    bool   `xorm:"bool 'scheduled' not null default false" json:"scheduled"`
	Deleted      bool   `xorm:"bool 'deleted' not null default false" json:"deleted"`
}
