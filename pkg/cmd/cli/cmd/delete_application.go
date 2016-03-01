package cmd

import (
	"errors"
	"fmt"
	applicationapi "github.com/openshift/origin/pkg/application/api"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"
	"io"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	errutil "k8s.io/kubernetes/pkg/util/errors"
)

const (
	deleteApplicationLong = `
application is a group of resource
delete application is used to delete resources in the same application
`
	deleteApplicationExample = `# delete a new application with [name deletelabel]
  $ %[1]s  mobile_app  --onlylabel=true
 `
)

type DeleteApplicationOptions struct {
	Name      string
	OnlyLabel bool

	Client  client.Interface
	KClient kclient.Interface

	Out io.Writer
}

func NewCmdDeleteApplication(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	options := &DeleteApplicationOptions{}
	options.Out = out

	cmd := &cobra.Command{
		Use:     `delete-application NAME [--onlylabel=true]`,
		Short:   "delete a existed application",
		Long:    deleteApplicationLong,
		Example: fmt.Sprintf(deleteApplicationExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			if err = options.complete(cmd, f); err != nil {
				kcmdutil.CheckErr(err)
				return
			}

			if options.Client, options.KClient, err = f.Clients(); err != nil {
				kcmdutil.CheckErr(err)
				return
			}

			if err := options.Run(f); err != nil {
				fmt.Println("run err %s", err.Error())
			}
		},
	}

	cmd.Flags().BoolVarP(&options.OnlyLabel, "onlylabel", "l", false, "only delete label")

	return cmd
}

func (o *DeleteApplicationOptions) complete(cmd *cobra.Command, f *clientcmd.Factory) error {
	args := cmd.Flags().Args()
	if len(args) != 1 {
		cmd.Help()
		return errors.New("must have exactly one argument")
	}

	return nil
}

func (o *DeleteApplicationOptions) Run(f *clientcmd.Factory) error {
	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	app, err := o.Client.Applications(namespace).Get(o.Name)
	if err != nil {
		return err
	}

	if !o.OnlyLabel {
		if errs := deleteAllContent(o.Client, o.KClient, app); len(errs) > 0 {
			return errutil.NewAggregate(errs)
		}
	}

	if err = o.Client.Applications(namespace).Delete(app.Name); err != nil {
		return err
	}

	return nil
}

func deleteAllContent(c client.Interface, kc kclient.Interface, app *applicationapi.Application) []error {
	errs := []error{}
	for _, item := range app.Spec.Items {
		switch item.Kind {
		case "Build":
			err := c.Builds(app.Namespace).Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "BuildConfig":
			err := c.BuildConfigs(app.Namespace).Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "DeploymentConfig":
			err := c.DeploymentConfigs(app.Namespace).Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "ImageStream":
			err := c.ImageStreams(app.Namespace).Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "ImageStreamTag":

		case "ImageStreamImage":

		case "Event":
			err := kc.Events(app.Namespace).Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "Node":
			err := kc.Nodes().Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "Job":

		case "Pod":
			// todo make sure deleteOption
			err := kc.Pods(app.Namespace).Delete(item.Name, nil)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "ReplicationController":
			err := kc.ReplicationControllers(app.Namespace).Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "Service":
			err := kc.Services(app.Namespace).Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "PersistentVolume":
			err := kc.PersistentVolumes().Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "PersistentVolumeClaim":
			err := kc.PersistentVolumeClaims(app.Namespace).Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "ServiceBroker":
			err := c.ServiceBrokers().Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "BackingService":
			err := c.BackingServices().Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		case "BackingServiceInstance":
			err := c.BackingServiceInstances(app.Namespace).Delete(item.Name)
			if err != nil && kerrors.IsNotFound(err) {
				errs = append(errs, err)
			}

		default:
			err := errors.New("unknown resource " + item.Kind + "=" + item.Name)
			errs = append(errs, err)
		}
	}

	return errs
}
