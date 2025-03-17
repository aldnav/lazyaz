package main

import "github.com/aldnav/lazyaz/pkg/azuredevops"

func isSameAsUser(name string, user *azuredevops.UserProfile) bool {
	return name == user.DisplayName || name == user.Username
}
