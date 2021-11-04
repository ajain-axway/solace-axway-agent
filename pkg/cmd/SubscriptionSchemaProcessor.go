package cmd

import (
	"fmt"
	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/apic"
	"github.com/Axway/agent-sdk/pkg/jobs"
	"github.com/Axway/agent-sdk/pkg/util/log"
	"github.com/solace-iot-team/solace-axway-agent/pkg/gateway"
	"github.com/solace-iot-team/solace-axway-agent/pkg/solace"
	"net/url"
	"strings"
)

// Publishes the Subscription Schema
type SubscriptionSchemaPublisherJob struct {
	jobs.Job
}

func (j *SubscriptionSchemaPublisherJob) Status() error {
	// continually called determining the status of any dependencies for the job
	// returning an error means the job should not be executed
	return nil
}

func (j *SubscriptionSchemaPublisherJob) Ready() bool {
	// called prior to executing the job the first time
	// return true when the job can begin execution, false otherwise
	return true
}

func (j *SubscriptionSchemaPublisherJob) Execute() error {
	// called each time the job should be executed
	// returning an error stops continuous jobs from executing
	log.Debugf("Registering SubscriptionSchema sol-schema-webhook-1")
	return apic.NewSubscriptionSchemaBuilder(agent.GetCentralClient()).
		SetName(solace.SolaceCallbackSubscriptionSchema).
		AddProperty(apic.NewSubscriptionSchemaPropertyBuilder().
			SetName(solace.SolaceCallback).
			IsString().
			SetDescription("Callback URL of this AsyncAPI").
			SetRequired()).
		AddProperty(apic.NewSubscriptionSchemaPropertyBuilder().
			SetName(solace.SolaceCallbackTrustedCNS).
			IsString().
			SetDescription("Trusted CN-Names (Comma-separated)")).
		AddProperty(apic.NewSubscriptionSchemaPropertyBuilder().
			SetName(solace.SolaceHttpMethod).
			IsString().
			SetEnumValues([]string{solace.SolaceHttpMethodPost, solace.SolaceHttpMethodPut}).
			SetDescription("HTTP-Method / Verb").
			SetRequired()).
		AddProperty(apic.NewSubscriptionSchemaPropertyBuilder().
			SetName(solace.SolaceInvocationOrder).
			IsString().
			SetEnumValues([]string{solace.SolaceInvocationOrderParallel, solace.SolaceInvocationOrderSerial}).
			SetDescription("Parallel or serial invocation of callback url").
			SetRequired()).
		AddProperty(apic.NewSubscriptionSchemaPropertyBuilder().
			SetName(solace.SolaceAuthenticationMethod).
			IsString().
			SetEnumValues([]string{solace.SolaceAuthenticationMethodNoAuthentication, solace.SolaceAuthenticationMethodBasicAuthentication, solace.SolaceAuthenticationMethodHttpHeaderAuthentication}).
			SetDescription("Authentication method").
			SetRequired()).
		AddProperty(apic.NewSubscriptionSchemaPropertyBuilder().
			SetName(solace.SolaceAuthenticationIdentifier).
			IsString().
			SetDescription("Authentication Username or Header Name")).
		AddProperty(apic.NewSubscriptionSchemaPropertyBuilder().
			SetName(solace.SolaceAuthenticationSecret).
			IsString().
			SetDescription("Authentication Password or Header Value")).
		Update(true).
		Register()
}

type SubscriptionSchemaProcessorJob struct {
	jobs.Job
}

func (j *SubscriptionSchemaProcessorJob) Status() error {
	// continually called determining the status of any dependencies for the job
	// returning an error means the job should not be executed
	return nil
}

func (j *SubscriptionSchemaProcessorJob) Ready() bool {
	// called prior to executing the job the first time
	// return true when the job can begin execution, false otherwise
	return true
}

func (j *SubscriptionSchemaProcessorJob) Execute() error {
	// called each time the job should be executed
	// returning an error stops continuous jobs from executing
	log.Tracef("SubscriptionSchemaProcessorJob triggered")
	resultlist, err := agent.GetCentralClient().GetApiServicesByQuery(solace.SolaceCallbackEnabledAttributeQuery)
	if err != nil {
		log.Errorf("SubscriptionSchemaProcessorJob: Could not query ApiServices (%s)", solace.SolaceCallbackEnabledAttributeQuery, err)
		return err
	} else {
		for _, service := range resultlist {
			log.Tracef("SubscriptionSchemaProcessorJob: Processing ApiService: %s ", service.Name)
			cq := fmt.Sprintf("metadata.references.kind==APIService and metadata.references.name==%s", service.Name)
			consumerInstances, errCi := agent.GetCentralClient().GetConsumerInstancesByQuery(cq)
			if errCi != nil {
				log.Errorf("SubscriptionSchemaProcessorJob:  Could not query ConsumerInstances", errCi)
				return errCi
			} else {
				for _, ci := range consumerInstances {
					log.Tracef("SubscriptionSchemaProcessorJob: Processing ConsumerInstance: %s ", ci.Name)
					if ci.Spec.Subscription.SubscriptionDefinition == solace.SolaceCallbackSubscriptionSchema {
						//nothing to do
					} else {
						errAttachSchema := agent.GetCentralClient().UpdateConsumerInstanceSubscriptionDefinitionByConsumerInstanceId(ci.Metadata.ID, solace.SolaceCallbackSubscriptionSchema)
						if errAttachSchema != nil {
							log.Errorf("SubscriptionSchemaProcessorJob: Could not attach Subscription Schema to ConsumerInstance:%s", ci.Name, errAttachSchema)
							return errAttachSchema
						} else {
							log.Infof("SubscriptionSchemaProcessorJob: Attached SubscriptionSchema: %s to ConsumerInstance: %s", solace.SolaceCallbackSubscriptionSchema, ci.Name)
						}
					}
				}
			}
		}
	}
	return nil
}

// validates Solace Callback attributes in subscription schema
func validateSolaceCallbackSubscription(subscription apic.Subscription) (bool, string) {
	log.Tracef(" Handling validateSubscription for [Subscription:%s] ", subscription.GetName())
	_, err := gateway.NewSubscriptionContainer(subscription)
	if err != nil {
		log.Errorf("Handling validateSubscription for [Subscription:%s] was not successful. [%s]", subscription.GetName(), err.Error())
		return false, "Subscription could not get validated."
	}
	method := subscription.GetPropertyValue(solace.SolaceHttpMethod)
	callback := subscription.GetPropertyValue(solace.SolaceCallback)
	authentication := subscription.GetPropertyValue(solace.SolaceAuthenticationMethod)
	//invocationOrder := subscription.GetPropertyValue(SolaceInvocationOrder)
	authenticationSecret := strings.TrimSpace(subscription.GetPropertyValue(solace.SolaceAuthenticationSecret))
	authenticationIdentifier := strings.TrimSpace(subscription.GetPropertyValue(solace.SolaceAuthenticationIdentifier))
	trustedCNS := strings.TrimSpace(subscription.GetPropertyValue(solace.SolaceCallbackTrustedCNS))
	validationFeedback := ""
	//is it a solace-callback subscription?
	// method is the indicator
	if len(method) > 0 {
		callbackUrl, err := url.ParseRequestURI(callback)
		if err != nil {
			validationFeedback = "Callback URL is invalid."
		} else {
			if strings.ToLower(callbackUrl.Scheme) == "https" {
				log.Tracef("IT IS A HTTPS CALLBACK with trusted CNs:%s", trustedCNS)

			} else {
				log.Tracef("IT IS A HTTP CALLBACK")
			}
		}
		if authentication == solace.SolaceAuthenticationMethodBasicAuthentication || authentication == solace.SolaceAuthenticationMethodHttpHeaderAuthentication {
			if len(authenticationIdentifier) == 0 {
				validationFeedback = fmt.Sprintf("%s %s", validationFeedback, "Username / Header Name is missing.")
			}
			if len(authenticationSecret) == 0 {
				validationFeedback = fmt.Sprintf("%s %s", validationFeedback, "Password / Header Value is missing.")
			}
		}
		if len(validationFeedback) > 0 {
			return false, validationFeedback
		}
	}
	return true, ""
}
