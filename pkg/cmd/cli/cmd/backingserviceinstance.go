package cmd

import (
	"errors"
	"fmt"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
	"github.com/spf13/cobra"
	"io"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	
	log "github.com/golang/glog"
)

//====================================================
// new
//====================================================

const (
	newBackingServiceInstanceLong = `
Create a new BackingServiceInstance

This command will try to create a backing service instance.
`
	newBackingServiceInstanceExample = `# Create a new backingserviceinstance with [name BackingServiceName BackingServicePlanGuid]
  $ %[1]s mysql_BackingServiceInstance --backingservice_name="BackingServiceName" --planid="BackingServicePlanGuid"`
)

type NewBackingServiceInstanceOptions struct {
	Name      string
	
	BackingServiceName     string
	BackingServicePlanGuid string

	Client client.Interface

	Out io.Writer
}

func NewCmdNewBackingServiceInstance(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &NewBackingServiceInstanceOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     "new-backingserviceinstance NAME --backingservice_name=BackingServiceName --planid=BackingServicePlanGuid",
		Short:   "create a new BackingServiceInstance",
		Long:    newBackingServiceInstanceLong,
		Example: fmt.Sprintf(newBackingServiceInstanceExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if options.complete(cmd, f); err != nil {
				kcmdutil.CheckErr(err)
			}

			if options.Client, _, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := options.Run(f); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.BackingServiceName, "backingservice_name", "", "BackingService GUID")
	cmd.Flags().StringVar(&options.BackingServicePlanGuid, "planid", "", "BackingService Plan GUID")
	// todo: dashboard_url
	
	return cmd
}

func (o *NewBackingServiceInstanceOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) == 0 {
		cmd.Help()
		return errors.New("must have at least 1 argument")
	}

	o.Name = args[0]

	return nil
}

func (o *NewBackingServiceInstanceOptions) Run(f *clientcmd.Factory) error {
	backingServiceInstance := &backingserviceinstanceapi.BackingServiceInstance{}
	
	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	
	backingServiceInstance.Annotations = make(map[string]string)
	backingServiceInstance.Name = o.Name
	backingServiceInstance.GenerateName = o.Name
	
	backingServiceInstance.Spec.Provisioning.BackingServiceName = o.BackingServiceName
	backingServiceInstance.Spec.Provisioning.BackingServicePlanGuid = o.BackingServicePlanGuid
	backingServiceInstance.Spec.Provisioning.Parameters = make(map[string]string)
	
	//backingServiceInstance.Spec.Binding.BindUuid = 
	backingServiceInstance.Spec.Binding.InstanceBindDeploymentConfig = make(map[string]string)
	backingServiceInstance.Spec.Binding.Credential = make(map[string]string)
	
	//backingServiceInstance.Status = 
	
	_, err = o.Client.BackingServiceInstances(namespace).Create(backingServiceInstance)
	if err != nil {
		return err
	}

	return nil
}

//====================================================
// edit
//====================================================

const (
	editBackingServiceInstanceLong = `
Edit a BackingServiceInstance

This command will try to edit a backing service instance.
`
	editBackingServiceInstanceExample = `# Edit a backingserviceinstance with [name BackingServicePlanGuid]
  $ %[1]s mysql_BackingServiceInstance --planid="BackingServicePlanGuid"`
)

type EditBackingServiceInstanceOptions struct {
	Name      string
	
	BackingServicePlanGuid string

	Client client.Interface

	Out io.Writer
}

func NewCmdEditBackingServiceInstance(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &NewBackingServiceInstanceOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     "edit-backingserviceinstance NAME --plan_guid=BackingServicePlanGuid",
		Short:   "Edit a BackingServiceInstance",
		Long:    editBackingServiceInstanceLong,
		Example: fmt.Sprintf(editBackingServiceInstanceExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if options.complete(cmd, f); err != nil {
				kcmdutil.CheckErr(err)
			}

			if options.Client, _, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := options.Run(f); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.BackingServicePlanGuid, "planid", "", "BackingService Plan GUID")
	
	return cmd
}

func (o *EditBackingServiceInstanceOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) == 0 {
		cmd.Help()
		return errors.New("must have at least 1 argument")
	}

	o.Name = args[0]

	return nil
}

func (o *EditBackingServiceInstanceOptions) Run(f *clientcmd.Factory) error {
	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	
	backingServiceInstance, err := o.Client.BackingServiceInstances(namespace).Get(o.Name)
	if err != nil {
		return err
	}
	
	backingServiceInstance.Spec.Provisioning.BackingServicePlanGuid = o.BackingServicePlanGuid
	
	_, err = o.Client.BackingServiceInstances(namespace).Update(backingServiceInstance)
	if err != nil {
		return err
	}

	return nil
}

//====================================================
// bind
//====================================================

const (
	bindBackingServiceInstanceLong = `
Bind a new BackingServiceInstance

This command will try to bind a backing service instance and a deployment config.
`
	bindBackingServiceInstanceExample = `# Bind a new backingserviceinstance with a deploy config [BackingServiceInstanceName DeploymentConfigName]
  $ %[1]s mysql_BackingServiceInstance helloworld_DeploymentConfig`
)

type BindBackingServiceInstanceOptions struct {
	Name                 string
	DeploymentConfigName string

	Client client.Interface

	Out io.Writer
}

func NewCmdBindBackingServiceInstance(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &BindBackingServiceInstanceOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     "bind-backingserviceinstance BackingServiceInstanceName DeployConfigName",
		Short:   "bind a BackingServiceInstance and a DeployConfig",
		Long:    bindBackingServiceInstanceLong,
		Example: fmt.Sprintf(bindBackingServiceInstanceExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if options.complete(cmd, f); err != nil {
				kcmdutil.CheckErr(err)
			}

			if options.Client, _, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := options.Run(f); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}
	
	return cmd
}

func (o *BindBackingServiceInstanceOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) < 2 {
		cmd.Help()
		return errors.New("must have at least 2 arguments")
	}

	o.Name                 = args[0]
	o.DeploymentConfigName = args[1]

	return nil
}

func (o *BindBackingServiceInstanceOptions) Run(f *clientcmd.Factory) error {
	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	
	bsi, err := o.Client.BackingServiceInstances(namespace).Get(o.Name)
	if err != nil {
		return err
	}
	
	dc, err := o.Client.DeploymentConfigs(namespace).Get(o.DeploymentConfigName)
	if err != nil {
		return err
	}
	
	log.Infoln("to bind bsi (", bsi.Name, " and (", dc.Name, ")")

	return nil
}

//====================================================
// unbind
//====================================================

const (
	unbindBackingServiceInstanceLong = `
Unbind a new BackingServiceInstance

This command will try to unbind a backing service instance and a deployment config.
`
	unbindBackingServiceInstanceExample = `# Unbind a new backingserviceinstance with and deploy config [BackingServiceInstanceName DeploymentConfigName]
  $ %[1]s mysql_BackingServiceInstance helloworld_DeploymentConfig`
)

type UnbindBackingServiceInstanceOptions struct {
	Name                 string
	DeploymentConfigName string

	Client client.Interface

	Out io.Writer
}

func NewCmdUnbindBackingServiceInstance(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &UnbindBackingServiceInstanceOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     "unbind-backingserviceinstance BackingServiceInstanceName DeployConfigName",
		Short:   "unbind a BackingServiceInstance and a DeployConfig",
		Long:    unbindBackingServiceInstanceLong,
		Example: fmt.Sprintf(unbindBackingServiceInstanceExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if options.complete(cmd, f); err != nil {
				kcmdutil.CheckErr(err)
			}

			if options.Client, _, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := options.Run(f); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}
	
	return cmd
}

func (o *UnbindBackingServiceInstanceOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) < 2 {
		cmd.Help()
		return errors.New("must have at least 2 arguments")
	}

	o.Name                 = args[0]
	o.DeploymentConfigName = args[1]

	return nil
}

func (o *UnbindBackingServiceInstanceOptions) Run(f *clientcmd.Factory) error {
	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	
	bsi, err := o.Client.BackingServiceInstances(namespace).Get(o.Name)
	if err != nil {
		return err
	}
	
	dc, err := o.Client.DeploymentConfigs(namespace).Get(o.DeploymentConfigName)
	if err != nil {
		return err
	}
	
	log.Infoln("to unbind bsi (", bsi.Name, " and (", dc.Name, ")")

	return nil
}


