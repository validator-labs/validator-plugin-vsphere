package constants

const (
	PluginCode string = "AWS"

	ValidationTypeIAMRolePolicy  string = "aws-iam-role-policy"
	ValidationTypeIAMUserPolicy  string = "aws-iam-user-policy"
	ValidationTypeIAMGroupPolicy string = "aws-iam-group-policy"
	ValidationTypeRolePrivileges string = "aws-iam-policy"
	ValidationTypeServiceQuota   string = "aws-service-quota"
	ValidationTypeTag            string = "aws-tag"

	IAMWildcard string = "*"
)
