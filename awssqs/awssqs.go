/*
 * http://www.apache.org/licenses/LICENSE-2.0.txt
 *
 * Copyright 2017 OpsVision Solutions
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package awssqs

// Imports
import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

// Constants
const (
	pluginVendor  = "opsvision"
	pluginName    = "awssqs"
	pluginVersion = 1
)

// AWSSQS
type AWSSQS struct {
	initialized bool
	service     *sqs.SQS
}

// New is our constructor
func New() *AWSSQS {
	return new(AWSSQS)
}

// initialize
func (s *AWSSQS) init(cfg plugin.Config) {
	s.initialized = true
}

// GetConfigPolicy - Returns the configPolicy for the plugin
func (s *AWSSQS) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	return *policy, nil
}

// Publish - Publishes metrics to AWSSQS using the TOKEN found in the config
func (s *AWSSQS) Publish(mts []plugin.Metric, cfg plugin.Config) error {
	return nil
}

// getAwsId obtains the AWS Key ID from the config file
func (s *AWSSQS) getAwsId(cfg plugin.Config) string {
	akid, err := cfg.GetString("akid")
	if err != nil {
		log.Fatalf("Error: Failed to find the 'akid' in config file\n")
	}

	return akid
}

// getAwsSecret obtains the AWS Secret from the config file
func (s *AWSSQS) getAwsSecret(cfg plugin.Config) string {
	secret, err := cfg.GetString("secret")
	if err != nil {
		log.Fatalf("Error: Failed to find 'secret' in config file\n")
	}

	return secret
}

// getAwsQueue obtains the AWS SQS Queue from the config file
func (s *AWSSQS) getAwsQueue(cfg plugin.Config) string {
	queue, err := cfg.GetString("queue")
	if err != nil {
		log.Fatalf("Error: Failed to find 'queue' in config file\n")
	}

	return queue
}

// extractRegion extracts the AWSSQS region from the supplied queue url. the
// complete list of AWSSQS regions can be found here:
//   http://docs.aws.amazon.com/general/latest/gr/rande.html#sqs_region
func (s *AWSSQS) extractRegion(queue string) string {
	// Setup our regular expression - queue's follow a pattern
	re := regexp.MustCompile(`sqs\.(.*?)\.amazonaws\.com`)

	// Check to see if our pattern exists
	if !re.MatchString(queue) {
		log.Fatalf("Error: Failed to extract region from queue - %s\n", queue)
	}

	// Extract the region from the queue url
	region := re.FindStringSubmatch(queue)[1]
	return region
}

// connect to the AWS SQS endpoint
func (s *AWSSQS) connect(akid string, secret string, queue string) {
	// Create credentials
	creds := credentials.NewStaticCredentials(akid, secret, "")

	// Creat the session
	region := s.extractRegion(queue)
	session := session.New(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})

	// Create the SQS service
	service := sqs.New(session)
	s.service = service
}
