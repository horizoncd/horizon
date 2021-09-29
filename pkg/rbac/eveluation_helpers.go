/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rbac

import "strings"

func VerbMatches(rule *PolicyRule, requestedVerb string) bool {
	for _, rulesVerb := range rule.Verbs {
		if rulesVerb == VerbAll {
			return true
		}
		if rulesVerb == requestedVerb {
			return true
		}
	}

	return false
}

func APIGroupMatches(rule *PolicyRule, requestedAPIGroup string) bool {
	for _, ruleAPIGroup := range rule.APIGroups {
		if ruleAPIGroup == APIGroupAll {
			return true
		}
		if ruleAPIGroup == requestedAPIGroup {
			return true
		}
	}

	return false
}

func ResourceMatches(rule *PolicyRule, combinedRequestedResource, requestedSubresource string) bool {
	for _, ruleResource := range rule.Resources {
		// if everything is allowed, we match
		if ruleResource == ResourceAll {
			return true
		}
		// if we have an exact match, we match
		if ruleResource == combinedRequestedResource {
			return true
		}

		// We can also match a */subresource.
		// if there  isn't a subresource, then continue
		if len(requestedSubresource) == 0 {
			continue
		}

		// if the rule isn't the format */subresource,
		// then we don't match, continue
		if len(ruleResource) == len(requestedSubresource)+2 &&
			strings.HasPrefix(ruleResource, "*/") &&
			strings.HasSuffix(ruleResource, requestedSubresource) {
			return true
		}
	}

	return false
}

func ScopeMatches(rule *PolicyRule, requestScope string) bool {
	for _, scopeRule := range rule.Scopes {
		if scopeRule == ScopeAll {
			return true
		}
		if scopeRule == requestScope {
			return true
		}
		if strings.HasSuffix(scopeRule, "*") &&
			strings.HasPrefix(requestScope, strings.TrimRight(scopeRule, "*")) {
			return true
		}
	}
	return false
}

func NonResourceURLMatches(rule *PolicyRule, requestedURL string) bool {
	for _, ruleURL := range rule.NonResourceURLs {
		if ruleURL == NonResourceAll {
			return true
		}
		if ruleURL == requestedURL {
			return true
		}
		if strings.HasSuffix(ruleURL, "*") &&
			strings.HasPrefix(requestedURL, strings.TrimRight(ruleURL, "*")) {
			return true
		}
	}
	return false
}
