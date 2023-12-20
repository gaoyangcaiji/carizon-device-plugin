package errors

const (
	CCSystemBusy         = -1
	CCSystemUnknownError = -2
	CCSuccess            = 0
	CCSuccessStr         = "success"
	CCNoPermission       = 9900403

	CCErrCommHTTPDoRequestFailed = 500
	// CCErrCommParamsInvalid parameter validation in the body is not pass
	CCErrCommParamsInvalid = 1199006
	// CCErrCommParamsNeedSet the parameter unassigned
	CCErrCommParamsNeedSet       = 1199010
	CCErrCommPageLimitIsExceeded = 1199059
)
