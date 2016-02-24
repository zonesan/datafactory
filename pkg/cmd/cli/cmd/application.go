package cmd

import (
	"errors"
	"fmt"
	applicationapi "github.com/openshift/origin/pkg/application/api"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
	"io"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	newApplicationLong = `
Create a new application to partition resources for a comfortable knowledge of my services

`
	newApplicationExample = `# Create a new application with [name items]
  $ %[1]s  mobile_app  --items="Pod=php,Pod=mysql,ServiceBrokerInstance=redis"`
)

type NewApplicationOptions struct {
	Name   string
	Items  string
	Client client.Interface

	Out io.Writer
}

func NewCmdApplication(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &NewApplicationOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     `new-application NAME [--items="KIND=KINDNAME,KIND=KINDNAME"]`,
		Short:   "create a new application",
		Long:    newApplicationLong,
		Example: fmt.Sprintf(newApplicationExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if err = options.complete(cmd, f); err != nil {
				kcmdutil.CheckErr(err)
			}

			if options.Client, _, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := options.Run(f); err != nil {
				fmt.Println("run err %s", err.Error())
			}
		},
	}

	cmd.Flags().StringVar(&options.Items, "items", "", "Application Items")

	return cmd
}

func (o *NewApplicationOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) != 1 {
		cmd.Help()
		return errors.New("must have exactly one argument")
	}

	o.Name = args[0]

	return nil
}

func (o *NewApplicationOptions) Run(f *clientcmd.Factory) error {
	application := &applicationapi.Application{}
	//application.Spec.ItemsName = o.Name
	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	application.Annotations = make(map[string]string)
	application.Name = o.Name
	application.GenerateName = o.Name
	application.Status.Phase = applicationapi.ApplicationNew

	if _, err = o.Client.Applications(namespace).Create(application); err != nil {
		return err
	}

	return nil
}
