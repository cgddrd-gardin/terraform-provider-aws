package glue_test

import "testing"

func TestAccAWSGlue_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"ResourcePolicy": {
			"basic":      testAccResourcePolicy_basic,
			"update":     testAccResourcePolicy_update,
			"disappears": testAccResourcePolicy_disappears,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
