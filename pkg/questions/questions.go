package questions

import (
	"gopkg.in/AlecAivazis/survey.v1"
)

var configQuestions = []*survey.Question{
	{
		Name:     "username",
		Prompt:   &survey.Input{Message: "Enter your username:"},
		Validate: survey.Required,
	},
	{
		Name: "password",
		Prompt: &survey.Password{
			Message: "Enter your password:",
		},
		Validate: survey.Required,
	},
}

type Answers struct {
	Username string
	Password string
}

func Ask() Answers {
	answers := Answers{}
	err := survey.Ask(configQuestions, &answers)
	if err != nil {
		panic(err)
	}

	return answers
}

func ShouldSave() bool {
	shouldSave := false
	prompt := &survey.Confirm{
		Message: "Would you like to save these credentials?",
	}
	survey.AskOne(prompt, &shouldSave, nil)
	return shouldSave
}
