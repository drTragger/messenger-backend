package utils

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/rest/api/v2010"
)

type SMSClient struct {
	Client       *twilio.RestClient
	FromPhoneNum string
}

func NewSMSClient() *SMSClient {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromPhoneNum := os.Getenv("TWILIO_PHONE_NUMBER")

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	return &SMSClient{
		Client:       client,
		FromPhoneNum: fromPhoneNum,
	}
}

func (s *SMSClient) SendSMS(to, message string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(s.FromPhoneNum)
	params.SetBody(message)

	_, err := s.Client.Api.CreateMessage(params)
	if err != nil {
		log.Printf("Failed to send SMS: %v\n", err)
		return err
	}

	log.Printf("SMS sent successfully to %s\n", to)
	return nil
}

func GenerateRandomCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = digits[rand.Intn(len(digits))]
	}
	return string(code)
}
