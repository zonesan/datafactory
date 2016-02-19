package cmd

import (
	"errors"
	"fmt"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	servicebrokerapi "github.com/openshift/origin/pkg/servicebroker/api"
	"github.com/spf13/cobra"
	"io"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	newServiceBrokerLong = `
Create a new servicebroker for administrator

`
	newServiceBrokerExample = `# Create a new servicebroker with [name username password url]
  $ %[1]s  mysql_servicebroker  --username="username"  --password="password" --url="url"`
)

type NewServiceBrokerOptions struct {
	Url      string
	Name     string
	UserName string
	Password string

	Client client.Interface

	Out io.Writer
}

func NewCmdServiceBroker(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &NewServiceBrokerOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     "new-servicebroker NAME [--username=USERNAME] [--password=PASSWORD] [--url=URL]",
		Short:   "create a new servicebroker",
		Long:    newServiceBrokerLong,
		Example: fmt.Sprintf(newServiceBrokerExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if options.complete(cmd, f); err != nil {
				kcmdutil.CheckErr(err)
			}

			if options.Client, _, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := options.Run(); err != nil {
				fmt.Println("run err %s", err.Error())
			}
		},
	}

	cmd.Flags().StringVar(&options.Url, "url", "", "ServiceBroker Url")
	//	cmd.Flags().StringVar(&options.Name, "name", "", "ServiceBroker Name")
	cmd.Flags().StringVar(&options.UserName, "username", "", "ServiceBroker username")
	cmd.Flags().StringVar(&options.Password, "password", "", "ServiceBroker Password")

	return cmd
}

func (o *NewServiceBrokerOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) == 0 {
		cmd.Help()
		return errors.New("must have exactly one argument")
	}

	o.Name = args[0]

	return nil
}

func (o *NewServiceBrokerOptions) Run() error {
	serviceBroker := &servicebrokerapi.ServiceBroker{}
	serviceBroker.Spec.Name = o.Name
	serviceBroker.Spec.Url = o.Url
	serviceBroker.Spec.UserName = o.UserName
	serviceBroker.Spec.Password = o.Password
	serviceBroker.Annotations = make(map[string]string)
	serviceBroker.Name = o.Name
	serviceBroker.GenerateName = o.Name
	serviceBroker.Status.Phase = servicebrokerapi.ServiceBrokerNew

	_, err := o.Client.ServiceBrokers().Create(serviceBroker)
	if err != nil {
		return err
	}

	return nil
}
