// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package job

type Config struct {

	// LockName is the name of the lock resource
	LockName string `yaml:"lockName"`

	// LockNS is the namespace of the lock resource
	LockNS string `yaml:"lockNS"`

	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership
	LeaseDuration int `yaml:"leaseDuration"`

	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up
	RenewDeadline int `yaml:"renewDeadline"`

	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions
	RetryPeriod int `yaml:"retryPeriod"`
}
