package authz

const (
	ResourceCompliance = "compliance"
)

var (
	ComplianceRead  = Permission{Resource: ResourceCompliance, Action: ActionRead}
	ComplianceWrite = Permission{Resource: ResourceCompliance, Action: ActionWrite}
)

func init() {
	RegisterResource(ResourceDefinition{
		Resource: ResourceCompliance,
		LabelKey: "Compliance",
		Actions: []ActionDefinition{
			{
				Action:         ActionRead,
				LabelKey:       "Read compliance records",
				DescriptionKey: "View privacy requests and public feedback records.",
				DefaultRoles:   []string{BuiltInRoleAdmin},
			},
			{
				Action:         ActionWrite,
				LabelKey:       "Handle compliance records",
				DescriptionKey: "Update privacy requests and public feedback handling status.",
				DefaultRoles:   []string{BuiltInRoleAdmin},
			},
		},
	})
}
