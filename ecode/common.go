package ecode

var (
	OK = add(0)
	ServerErr = add(500)
	TooManyRequest = add(429)

	NoSuchActivity = add(10001)
	AddLotteryQuotaReachLimit = add(10002)
	NoEnoughDrawQuota = add(10003)
	NoSuchAward = add(10004)
	LotteryIsTooHot = add(10005)
	PetalExchangeFailed = add(10006)
	CaseReportIsNotStarted = add(11000)
	CaseReportIsEnd        = add(11001)
)
