package cmd

import (
	"errors"
	"fmt"
	applicationapi "github.com/openshift/origin/pkg/application/api"
	applicationutil "github.com/openshift/origin/pkg/application/util"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
	"io"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"strings"
)

const (
	newApplicationLong = `
Create a new application to partition resources for a comfortable knowledge of my services

`
	newApplicationExample = `# Create a new application with [name items]
  $ %[1]s  mobile_app  --items="Pod=php,Pod=mysql,ServiceBrokerInstance=redis"`
)

type NewApplicationOptions struct {
	Name  string
	Items applicationapi.ItemList
	Item string

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
				return
			}

			if options.Client, _, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := options.Run(f); err != nil {
				fmt.Println("run err %s", err.Error())
			}
		},
	}

	cmd.Flags().StringVar(&options.Item, "items", "", "application items")

	return cmd
}

func (o *NewApplicationOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) != 1 {
		cmd.Help()
		return errors.New("must have exactly one argument")
	}

	flagItems := strings.TrimSpace(o.Item)
	if len(flagItems) == 0 {
		return errors.New("items length must not be 0")
	}

	items, err := applicationutil.Parse(flagItems)
	if err != nil {
		return err
	}

	o.Items = items
	o.Name = args[0]

	return nil
}

func (o *NewApplicationOptions) Run(f *clientcmd.Factory) error {
	application := &applicationapi.Application{}

	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	application.Spec.Items = o.Items
	application.Annotations = make(map[string]string)
	application.Labels = map[string]string{}
	application.Name = o.Name
	application.GenerateName = o.Name
	application.Status.Phase = applicationapi.ApplicationNew

	if _, err = o.Client.Applications(namespace).Create(application); err != nil {
		return err
	}

	return nil
}
