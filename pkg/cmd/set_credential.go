package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ZupIT/ritchie-cli/pkg/credential"
	"github.com/ZupIT/ritchie-cli/pkg/credential/credsingle"
	"github.com/ZupIT/ritchie-cli/pkg/prompt"
	"github.com/ZupIT/ritchie-cli/pkg/stdin"
)

var inputTypes = []string{"plain text", "secret"}

// setCredentialCmd type for set credential command
type setCredentialCmd struct {
	credential.Setter
	credential.SingleSettings
	prompt.InputText
	prompt.InputBool
	prompt.InputList
	prompt.InputPassword
}

// NewSetCredentialCmd creates a new cmd instance
func NewSetCredentialCmd(
	credSetter credential.Setter,
	credSetting credential.SingleSettings,
	inText prompt.InputText,
	inBool prompt.InputBool,
	inList prompt.InputList,
	inPass prompt.InputPassword,
) *cobra.Command {
	s := &setCredentialCmd{
		Setter:         credSetter,
		SingleSettings: credSetting,
		InputText:      inText,
		InputBool:      inBool,
		InputList:      inList,
		InputPassword:  inPass,
	}

	cmd := &cobra.Command{
		Use:   "credential",
		Short: "Set credential",
		Long:  `Set credentials for Github, Gitlab, AWS, UserPass, etc.`,
		RunE:  RunFuncE(s.runStdin(), s.runPrompt()),
	}
	cmd.LocalFlags()

	return cmd
}

func (s setCredentialCmd) runPrompt() CommandRunnerFunc {
	return func(cmd *cobra.Command, args []string) error {
		cred, err := s.prompt()
		if err != nil {
			return err
		}

		if err := s.Set(cred); err != nil {
			return err
		}

		prompt.Success(fmt.Sprintf("✔ %s credential saved!", strings.Title(cred.Service)))
		return nil
	}
}

func (s setCredentialCmd) prompt() (credential.Detail, error) {

	if err := s.WriteDefaultCredentials(credsingle.ProviderPath()); err != nil {
		return credential.Detail{}, err
	}

	var credDetail credential.Detail
	cred := credential.Credential{}

	credentials, err := s.ReadCredentials(credsingle.ProviderPath())
	if err != nil {
		return credential.Detail{}, err
	}

	providerArr := credsingle.NewProviderArr(credentials)
	providerChoose, err := s.List("Select your provider", providerArr)
	if err != nil {
		return credDetail, err
	}

	if providerChoose == credsingle.AddNew {
		newProvider, err := s.Text("Define your provider name:", true)
		if err != nil {
			return credDetail, err
		}
		providerArr = append(providerArr, newProvider)

		var newFields []credential.Field
		var newField credential.Field
		addMoreCredentials := true
		for addMoreCredentials {
			newField.Name, err = s.Text("Define your field name: (ex.:token, secretAccessKey)", true)
			if err != nil {
				return credDetail, err
			}

			newField.Type, err = s.List("Select your field type:", inputTypes)
			if err != nil {
				return credDetail, err
			}

			newFields = append(newFields, newField)
			addMoreCredentials, err = s.Bool("Add more credentials to this provider?", []string{"no", "yes"})
			if err != nil {
				return credDetail, err
			}
		}
		credentials[newProvider] = newFields
		if err = s.WriteCredentials(credentials, credsingle.ProviderPath()); err != nil {
			return credDetail, err
		}

		providerChoose = newProvider
	}

	inputs := credentials[providerChoose]

	for _, i := range inputs {
		var value string
		if i.Type == inputTypes[1] {
			value, err = s.Password(i.Name + ":")
			if err != nil {
				return credDetail, err
			}
		} else {
			value, err = s.Text(i.Name+":", true)
			if err != nil {
				return credDetail, err
			}
		}
		cred[i.Name] = value
	}
	credDetail.Service = providerChoose
	credDetail.Credential = cred

	return credDetail, nil
}

func (s setCredentialCmd) runStdin() CommandRunnerFunc {
	return func(cmd *cobra.Command, args []string) error {
		cred, err := s.stdinResolver()
		if err != nil {
			return err
		}

		if err := s.Set(cred); err != nil {
			return err
		}

		prompt.Success(fmt.Sprintf("✔ %s credential saved!", strings.Title(cred.Service)))
		return nil
	}
}

func (s setCredentialCmd) stdinResolver() (credential.Detail, error) {
	var credDetail credential.Detail

	if err := stdin.ReadJson(os.Stdin, &credDetail); err != nil {
		prompt.Error(stdin.MsgInvalidInput)
		return credDetail, err
	}
	return credDetail, nil
}
