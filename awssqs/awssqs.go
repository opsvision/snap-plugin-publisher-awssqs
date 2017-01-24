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
	//"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

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

// AWSSQS is our main class and is used to hold useful properties
type AWSSQS struct {
	initialized bool
	hostname    string
	queue       string
	service     *sqs.SQS
}

// New is our constructor
func New() *AWSSQS {
	return new(AWSSQS)
}

// initialize
func (s *AWSSQS) init(cfg plugin.Config) {
	if s.initialized {
		return
	}

	// Enable debugging if requested - call this first!
	s.setDebugFile(cfg)
	log.Printf("AWSSQS Plugin Initialized")

	// Get the hostname
	s.getHostname()

	// Get our required configuration parameters
	akid := s.getAwsID(cfg)
	secret := s.getAwsSecret(cfg)
	queue := s.getAwsQueue(cfg)
	s.queue = queue

	// Connect to Amazon
	s.connect(akid, secret)

	s.initialized = true
}

// GetConfigPolicy - Returns the configPolicy for the plugin
func (s *AWSSQS) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	// The AWS API Key ID
	policy.AddNewStringRule([]string{pluginVendor, pluginName},
		"akid",
		true)

	// The AWS Secret
	policy.AddNewStringRule([]string{pluginVendor, pluginName},
		"secret",
		true)

	// The AWS SQS queue url
	policy.AddNewStringRule([]string{pluginVendor, pluginName},
		"queue",
		true)

	// The file name to use when debugging (optional)
	policy.AddNewStringRule([]string{pluginVendor, pluginName},
		"debug-file",
		false)

	return *policy, nil
}

// Publish - Publishes metrics to AWSSQS using the TOKEN found in the config
func (s *AWSSQS) Publish(mts []plugin.Metric, cfg plugin.Config) error {
	if len(mts) > 0 {
		s.init(cfg)
	}

	// Get the current time
	t := time.Now()

	// iterate over the incoming metrics
	for _, m := range mts {
		// create our message
		msg := map[string]string{
			"hostname":  s.hostname,
			"plugin":    pluginName,
			"metric":    strings.Join(m.Namespace.Strings(), "."),
			"value":     fmt.Sprintf("%v", m.Data),
			"type":      fmt.Sprintf("%s", reflect.TypeOf(m.Data)),
			"timestamp": fmt.Sprintf("%s", t.Format(time.RFC3339)),
		}

		// convert the message to json
		json, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("Failed to marshall %v", msg)
		}

		// send the message
		_, err = s.service.SendMessage(&sqs.SendMessageInput{
			QueueUrl:    aws.String(s.queue),
			MessageBody: aws.String(string(json)),
		})

		// log errors
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

// getHostname attempts to determine the hostname
func (s *AWSSQS) getHostname() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	s.hostname = hostname
}

// setDebugFile will log to the specific debug_file in the config if present
func (s *AWSSQS) setDebugFile(cfg plugin.Config) {
	fileName, err := cfg.GetString("debug_file")
	if err != nil {
		//fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return
	}

	// Open the output file
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	// Set logging output for debugging
	log.SetOutput(f)
}

// getAwsID obtains the AWS Key ID from the config file
func (s *AWSSQS) getAwsID(cfg plugin.Config) string {
	log.Printf("Reading AWS Key ID ('akid') from Config\n")

	akid, err := cfg.GetString("akid")
	if err != nil {
		log.Fatalf("Error: Failed to find the 'akid' in config file\n")
	}

	return akid
}

// getAwsSecret obtains the AWS Secret from the config file
func (s *AWSSQS) getAwsSecret(cfg plugin.Config) string {
	log.Printf("Reading AWS Secret ('secret') from Config\n")

	secret, err := cfg.GetString("secret")
	if err != nil {
		log.Fatalf("Error: Failed to find 'secret' in config file\n")
	}

	return secret
}

// getAwsQueue obtains the AWS SQS Queue from the config file
func (s *AWSSQS) getAwsQueue(cfg plugin.Config) string {
	log.Printf("Reading AWS SQS Queue ('queue') from Config\n")

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
	log.Printf("Extracting region from %s\n", queue)

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
func (s *AWSSQS) connect(akid string, secret string) {
	log.Printf("Connecting to Amazon\n")

	// Create credentials
	creds := credentials.NewStaticCredentials(akid, secret, "")

	// Creat the session
	region := s.extractRegion(s.queue)
	session := session.New(&aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	})

	// Create the SQS service
	service := sqs.New(session)
	s.service = service
}
