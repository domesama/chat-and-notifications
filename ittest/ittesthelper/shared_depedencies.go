package ittesthelper

import (
	"fmt"
	"os"
	"strings"

	"github.com/domesama/chat-and-notifications/connections/connectionconfig"
)

func SuffixKafkaTopicName(suiteName string) connectionconfig.KafkaConsumerInfo {
	pid := os.Getpid()

	sanitizedName := strings.Split(suiteName, "/")[0]
	lowerCaseName := strings.ToLower(sanitizedName)

	return connectionconfig.KafkaConsumerInfo{
		ConsumerName: fmt.Sprintf("%s_consumer_%d", lowerCaseName, pid),
		TopicName:    fmt.Sprintf("%s_topic_%d", lowerCaseName, pid),
	}
}
