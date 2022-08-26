package lib

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	experimentTypes "litmus-chaos-toolkit/pkg/chaos-toolkit/types"
	clients "litmus-chaos-toolkit/pkg/clients"
	"litmus-chaos-toolkit/pkg/events"
	"litmus-chaos-toolkit/pkg/log"
	"litmus-chaos-toolkit/pkg/result"
	"litmus-chaos-toolkit/pkg/types"
	"litmus-chaos-toolkit/pkg/k8s"
	"litmus-chaos-toolkit/pkg/utils/common"
	"os"
	"os/exec"
	"time"
)

//PrepareChaosToolkit contains the prepration steps before chaos injection
func PrepareChaosToolkit(experimentsDetails *experimentTypes.ExperimentDetails, clients clients.ClientSets, resultDetails *types.ResultDetails, eventsDetails *types.EventDetails, chaosDetails *types.ChaosDetails) error {
	var err error
	//Waiting for the ramp time before chaos injection
	if experimentsDetails.RampTime != 0 {
		log.Infof("[Ramp]: Waiting for the %vs ramp time before injecting chaos", experimentsDetails.RampTime)
		common.WaitForDuration(experimentsDetails.RampTime)
	}

	// Getting the serviceAccountName, need permission inside helper pod to create the events
	if experimentsDetails.ChaosServiceAccount == "" {
		experimentsDetails.ChaosServiceAccount, err = common.GetServiceAccount(experimentsDetails.ChaosNamespace, experimentsDetails.ChaosPodName, clients)
		if err != nil {
			return errors.Errorf("unable to get the serviceAccountName, err: %v", err)
		}
	}


	err = k8s.LoadSecretAsEnv(experimentsDetails.CredentialsSecretName, experimentsDetails.ChaosNamespace)
        if err != nil {
                 return errors.Errorf("unable to load secret %s/%s as env variable, err: %v", experimentsDetails.ChaosNamespace, experimentsDetails.CredentialsSecretName, err)
        }

	err = execChaosToolkit(experimentsDetails, clients, eventsDetails, chaosDetails, resultDetails)
	if err != nil {
		return errors.Errorf("unable to exec chaos-toolkit, err: %v", err)
	}

	//Waiting for the ramp time after chaos injection
	if experimentsDetails.RampTime != 0 {
		log.Infof("[Ramp]: Waiting for the %vs ramp time after injecting chaos", experimentsDetails.RampTime)
		common.WaitForDuration(experimentsDetails.RampTime)
	}
	return nil
}

func execChaosToolkit(experimentsDetails *experimentTypes.ExperimentDetails, clients clients.ClientSets, eventsDetails *types.EventDetails, chaosDetails *types.ChaosDetails, resultDetails *types.ResultDetails) error {

	//ChaosStartTimeStamp contains the start timestamp, when the chaos injection begin
	ChaosStartTimeStamp := time.Now()
	duration := int(time.Since(ChaosStartTimeStamp).Seconds())

	for duration < experimentsDetails.ChaosDuration {

		log.InfoWithValues("[Info]: Details of application under chaos injection", logrus.Fields{
			"\nAppName":   experimentsDetails.AppName,
			"\nOrgName":   experimentsDetails.OrgName,
			"\nSpaceName": experimentsDetails.SpaceName,
		})

		// record the event inside chaosengine
		if experimentsDetails.EngineName != "" {
			msg := "Injecting " + experimentsDetails.ExperimentName + " chaos on application pod"
			types.SetEngineEventAttributes(eventsDetails, types.ChaosInject, msg, "Normal", chaosDetails)
			events.GenerateEvents(eventsDetails, clients, chaosDetails, "ChaosEngine")
		}

		if err := execChaosToolkitCmd(experimentsDetails.ExperimentName); err != nil {
			return err
		}
		//Waiting for the chaos interval after chaos injection
		if experimentsDetails.ChaosInterval != 0 {
			log.Infof("[Wait]: Wait for the chaos interval %vs", experimentsDetails.ChaosInterval)
			common.WaitForDuration(experimentsDetails.ChaosInterval)
		}

		duration = int(time.Since(ChaosStartTimeStamp).Seconds())
	}
	if err := result.AnnotateChaosResult(resultDetails.Name, chaosDetails.ChaosNamespace, "targeted", "pod", experimentsDetails.AppName); err != nil {
		return err
	}
	log.Infof("[Completion]: %v chaos has been completed", experimentsDetails.ExperimentName)
	return nil
}

//stopDockerContainer kill the application container
func execChaosToolkitCmd(experimentName string) error {
	//  var errOut bytes.Buffer
	cmd := exec.Command("chaos", "run", experimentName+".json")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return errors.Errorf("Unable to run command, err: %v; error output: %v", err /*, errOut.String()*/)
	}
	return nil
}
