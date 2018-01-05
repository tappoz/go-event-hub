package eventhub

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	testEhNamespace   = os.Getenv("EH_TEST_NAMESPACE")
	testEhName        = os.Getenv("EH_TEST_NAME")
	testSasPolicyName = os.Getenv("EH_TEST_SAS_POLICY_NAME")
	testSasPolicyKey  = os.Getenv("EH_TEST_SAS_POLICY_KEY")
	// the consumer group is for the receiver app to consume messages
	testConsumerGroupName = os.Getenv("EH_TEST_CONSUMER_GROUP")
)

// go test -v --run TestEndToEndAmqpMessageFlowUsingLatestKeywordForPartitionOffsetsAsFilter ./eventhub/
func TestEndToEndAmqpMessageFlowUsingLatestKeywordForPartitionOffsetsAsFilter(t *testing.T) {
	// create the AMQP receiver
	ehReceiver, err := NewReceiver(ReceiverOpts{
		EventHubNamespace: testEhNamespace,
		EventHubName:      testEhName,
		SasPolicyName:     testSasPolicyName,
		SasPolicyKey:      testSasPolicyKey,
		ConsumerGroupName: testConsumerGroupName,
		PartitionOffsets:  []string{LatestOffset, LatestOffset},
		// to make sure this test does not mess up the default path and the values within
		PartitionOffsetsPath: "/tmp/end_to_end_integration_test_offsets.csv",
		OffsetsFlushInterval: 800 * time.Millisecond,
		TokenExpiryInterval:  20 * time.Second,
		Debug:                true,
	})
	if err != nil {
		panic(err)
	}
	defer ehReceiver.Close()
	go func(r Receiver) {
		for err := range r.ErrorChan() {
			if err != nil {
				panic(err)
			}
		}
	}(ehReceiver)

	// create the AMQP sender
	ehSender, err := NewSender(SenderOpts{
		EventHubNamespace:   testEhNamespace,
		EventHubName:        testEhName,
		SasPolicyName:       testSasPolicyName,
		SasPolicyKey:        testSasPolicyKey,
		TokenExpiryInterval: 20 * time.Second,
		Debug:               true,
	})
	if err != nil {
		panic(err)
	}
	defer ehSender.Close()
	go func(s Sender) {
		for err := range s.ErrorChan() {
			if err != nil {
				panic(err)
			}
		}
	}(ehSender)

	// prepare the message
	thisMessage := fmt.Sprintf("FOO_BAR_BAZ, created at: %s", time.Now())

	// send a message (ignore the returned unique ID)
	_, err = ehSender.Send(thisMessage)
	if err != nil {
		t.Errorf("There has been an error sending '%v', the error message is: %v\n", thisMessage, err)
	}

	// trigger the async receiving of messages
	ehReceiver.AsyncFetch()

	currEhMsg := <-ehReceiver.ReceiveChan()

	// assert on the received message
	if currEhMsg.Body != thisMessage {
		t.Errorf("Expected to receive this message: '%s' but received this message instead: '%s'", thisMessage, currEhMsg)
	}
}