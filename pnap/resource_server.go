package pnap

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/PNAP/go-sdk-helper-bmc/command/bmcapi/server"
	"github.com/PNAP/go-sdk-helper-bmc/command/ipapi/ipblock"
	"github.com/PNAP/go-sdk-helper-bmc/receiver"

	bmcapiclient "github.com/phoenixnap/go-sdk-bmc/bmcapi"
)

const (
	pnapRetryTimeout       = 100 * time.Minute
	pnapDeleteRetryTimeout = 15 * time.Minute
	pnapRetryDelay         = 5 * time.Second
	pnapRetryMinTimeout    = 3 * time.Second
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(pnapRetryTimeout),
			Update: schema.DefaultTimeout(pnapRetryTimeout),
			Delete: schema.DefaultTimeout(pnapDeleteRetryTimeout),
		},

		Schema: map[string]*schema.Schema{
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"private_ip_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"public_ip_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"os": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ssh_keys": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cpu": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cores_per_cpu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cpu_frequency_in_ghz": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"action": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"install_default_ssh_keys": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ssh_key_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"reservation_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"pricing_model": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"rdp_allowed_ips": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"management_ui_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_password": {
				Type:     schema.TypeString,
				Computed: true,
				//Sensitive: true,
			},
			"management_access_allowed_ips": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"provisioned_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag_assignment": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  nil,
									},
									"is_billing_tag": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"created_by": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"network_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway_address": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"private_network_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"gateway_address": {
										Type:       schema.TypeString,
										Optional:   true,
										Computed:   true,
										Deprecated: "Deprecated in favour of a common gateway address across all networks available under network_configuration",
									},
									"configuration_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"private_networks": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"server_private_network": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"id": {
																Type:     schema.TypeString,
																Required: true,
															},
															"ips": {
																Type:     schema.TypeSet,
																Optional: true,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"dhcp": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
																Default:  nil,
															},
															"status_description": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"ip_blocks_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"configuration_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"ip_blocks": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"server_ip_block": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"id": {
																Type:     schema.TypeString,
																Required: true,
															},
															"vlan_id": {
																Type:     schema.TypeInt,
																Optional: true,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
									"relinquish_ip_blocks": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"public_network_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"public_networks": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"server_public_network": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"id": {
																Type:     schema.TypeString,
																Required: true,
															},
															"ips": {
																Type:     schema.TypeSet,
																Required: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"status_description": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {

	client := m.(receiver.BMCSDK)

	request := &bmcapiclient.ServerCreate{}
	request.Hostname = d.Get("hostname").(string)
	var desc = d.Get("description").(string)
	if len(desc) > 0 {
		request.Description = &desc
	}
	request.Os = d.Get("os").(string)
	request.Type = d.Get("type").(string)
	request.Location = d.Get("location").(string)
	var networkType = d.Get("network_type").(string)

	if len(networkType) > 0 {
		request.NetworkType = &networkType
	}

	var resId = d.Get("reservation_id").(string)
	if len(resId) > 0 {
		request.ReservationId = &resId
	}

	var prModel = d.Get("pricing_model").(string)
	if len(prModel) > 0 {
		request.PricingModel = &prModel
	}

	var installDefault = d.Get("install_default_ssh_keys").(bool)
	request.InstallDefaultSshKeys = &installDefault
	temp := d.Get("ssh_keys").(*schema.Set).List()
	keys := make([]string, len(temp))
	for i, v := range temp {
		keys[i] = fmt.Sprint(v)
	}
	//todo
	request.SshKeys = &keys

	temp1 := d.Get("ssh_key_ids").(*schema.Set).List()
	keyIds := make([]string, len(temp1))
	for i, v := range temp1 {
		keyIds[i] = fmt.Sprint(v)
	}
	//todo
	request.SshKeyIds = &keyIds

	temp2 := d.Get("rdp_allowed_ips").(*schema.Set).List()
	allowedIps := make([]string, len(temp2))
	for i, v := range temp2 {
		allowedIps[i] = fmt.Sprint(v)
	}
	temp3 := d.Get("management_access_allowed_ips").(*schema.Set).List()
	managementAccessAllowedIps := make([]string, len(temp3))
	for i, v := range temp3 {
		managementAccessAllowedIps[i] = fmt.Sprint(v)
	}
	if len(temp2) > 0 || len(temp3) > 0 {
		dtoOsConfiguration := bmcapiclient.OsConfiguration{}

		if len(temp2) > 0 {
			dtoWindows := bmcapiclient.OsConfigurationWindows{}
			dtoWindows.RdpAllowedIps = &allowedIps
			dtoOsConfiguration.Windows = &dtoWindows
		}
		if len(temp3) > 0 {
			dtoOsConfiguration.ManagementAccessAllowedIps = &managementAccessAllowedIps
		}
		request.OsConfiguration = &dtoOsConfiguration
	}
	tags := d.Get("tags").([]interface{})
	if len(tags) > 0 {
		tagsObject := make([]bmcapiclient.TagAssignmentRequest, len(tags))
		for i, j := range tags {
			tarObject := bmcapiclient.TagAssignmentRequest{}
			tagsItem := j.(map[string]interface{})

			tagAssign := tagsItem["tag_assignment"].([]interface{})[0]
			tagAssignItem := tagAssign.(map[string]interface{})

			tarObject.Name = tagAssignItem["name"].(string)
			value := tagAssignItem["value"].(string)
			if len(value) > 0 {
				tarObject.Value = &value
			}
			tagsObject[i] = tarObject
		}
		request.Tags = &tagsObject
	}

	// network block
	if d.Get("network_configuration") != nil && len(d.Get("network_configuration").([]interface{})) > 0 {

		networkConfiguration := d.Get("network_configuration").([]interface{})[0]
		networkConfigurationItem := networkConfiguration.(map[string]interface{})

		networkConfigurationObject := bmcapiclient.NetworkConfiguration{}
		gatewayAddress := networkConfigurationItem["gateway_address"].(string)
		if len(gatewayAddress) > 0 {
			networkConfigurationObject.GatewayAddress = &gatewayAddress
		}
		if networkConfigurationItem["private_network_configuration"] != nil && len(networkConfigurationItem["private_network_configuration"].([]interface{})) > 0 {
			privateNetworkConfiguration := networkConfigurationItem["private_network_configuration"].([]interface{})[0]
			privateNetworkConfigurationItem := privateNetworkConfiguration.(map[string]interface{})

			gatewayAddress := privateNetworkConfigurationItem["gateway_address"].(string)
			configurationType := privateNetworkConfigurationItem["configuration_type"].(string)
			privateNetworks := privateNetworkConfigurationItem["private_networks"].([]interface{})

			if len(gatewayAddress) > 0 || len(configurationType) > 0 || len(privateNetworks) > 0 {
				privateNetworkConfigurationObject := bmcapiclient.PrivateNetworkConfiguration{}
				if len(gatewayAddress) > 0 {
					privateNetworkConfigurationObject.GatewayAddress = &gatewayAddress
				}

				if len(configurationType) > 0 {
					privateNetworkConfigurationObject.ConfigurationType = &configurationType
				}

				networkConfigurationObject.PrivateNetworkConfiguration = &privateNetworkConfigurationObject
				if len(privateNetworks) > 0 {

					serPrivateNets := make([]bmcapiclient.ServerPrivateNetwork, len(privateNetworks))

					for k, j := range privateNetworks {
						serverPrivateNetworkObject := bmcapiclient.ServerPrivateNetwork{}

						privateNetworkItem := j.(map[string]interface{})

						serverPrivateNetwork := privateNetworkItem["server_private_network"].([]interface{})[0]
						serverPrivateNetworkItem := serverPrivateNetwork.(map[string]interface{})

						id := serverPrivateNetworkItem["id"].(string)
						tempIps := serverPrivateNetworkItem["ips"].(*schema.Set).List()

						NetIps := make([]string, len(tempIps))
						for i, v := range tempIps {
							NetIps[i] = fmt.Sprint(v)
						}
						dhcp := serverPrivateNetworkItem["dhcp"].(bool)

						if (len(id)) > 0 {
							serverPrivateNetworkObject.Id = id
						}
						if (len(NetIps)) > 0 {
							serverPrivateNetworkObject.Ips = &NetIps
						}

						serverPrivateNetworkObject.Dhcp = &dhcp
						serPrivateNets[k] = serverPrivateNetworkObject

					}
					privateNetworkConfigurationObject.PrivateNetworks = &serPrivateNets
				}
			}
		}
		if networkConfigurationItem["ip_blocks_configuration"] != nil && len(networkConfigurationItem["ip_blocks_configuration"].([]interface{})) > 0 {
			ipBlocksConfiguration := networkConfigurationItem["ip_blocks_configuration"].([]interface{})[0]
			ipBlocksConfigurationItem := ipBlocksConfiguration.(map[string]interface{})

			confType := ipBlocksConfigurationItem["configuration_type"].(string)
			ipBlocks := ipBlocksConfigurationItem["ip_blocks"].([]interface{})

			if len(confType) > 0 || len(ipBlocks) > 0 {
				ipBlocksConfigurationObject := bmcapiclient.IpBlocksConfiguration{}
				if len(confType) > 0 {
					ipBlocksConfigurationObject.ConfigurationType = &confType
				}

				networkConfigurationObject.IpBlocksConfiguration = &ipBlocksConfigurationObject
				if len(ipBlocks) > 0 {

					serIpBlocks := make([]bmcapiclient.ServerIpBlock, len(ipBlocks))

					for k, j := range ipBlocks {
						serverIpBlockObject := bmcapiclient.ServerIpBlock{}

						ipBlockItem := j.(map[string]interface{})

						serverIpBlock := ipBlockItem["server_ip_block"].([]interface{})[0]
						serverIpBlockItem := serverIpBlock.(map[string]interface{})

						id := serverIpBlockItem["id"].(string)
						vlanId := int32(serverIpBlockItem["vlan_id"].(int))

						if (len(id)) > 0 {
							serverIpBlockObject.Id = id
						}
						serverIpBlockObject.VlanId = &vlanId
						serIpBlocks[k] = serverIpBlockObject
					}
					ipBlocksConfigurationObject.IpBlocks = &serIpBlocks
				}
			}
		}
		if networkConfigurationItem["public_network_configuration"] != nil && len(networkConfigurationItem["public_network_configuration"].([]interface{})) > 0 {
			publicNetworkConfiguration := networkConfigurationItem["public_network_configuration"].([]interface{})[0]
			publicNetworkConfigurationItem := publicNetworkConfiguration.(map[string]interface{})
			publicNetworks := publicNetworkConfigurationItem["public_networks"].([]interface{})

			if len(publicNetworks) > 0 {
				publicNetworkConfigurationObject := bmcapiclient.PublicNetworkConfiguration{}
				networkConfigurationObject.PublicNetworkConfiguration = &publicNetworkConfigurationObject
				serPublicNets := make([]bmcapiclient.ServerPublicNetwork, len(publicNetworks))

				for k, j := range publicNetworks {
					serverPublicNetworkObject := bmcapiclient.ServerPublicNetwork{}

					publicNetworkItem := j.(map[string]interface{})

					serverPublicNetwork := publicNetworkItem["server_public_network"].([]interface{})[0]
					serverPublicNetworkItem := serverPublicNetwork.(map[string]interface{})

					id := serverPublicNetworkItem["id"].(string)
					tempIps := serverPublicNetworkItem["ips"].(*schema.Set).List()

					NetIps := make([]string, len(tempIps))
					for i, v := range tempIps {
						NetIps[i] = fmt.Sprint(v)
					}
					if (len(id)) > 0 {
						serverPublicNetworkObject.Id = id
					}
					if (len(NetIps)) > 0 {
						serverPublicNetworkObject.Ips = NetIps
					}
					serPublicNets[k] = serverPublicNetworkObject
				}
				publicNetworkConfigurationObject.PublicNetworks = &serPublicNets
			}
		}
		request.NetworkConfiguration = &networkConfigurationObject
		b, _ := json.MarshalIndent(request, "", "  ")
		log.Printf("request object is" + string(b))
	}

	// end of network block
	requestCommand := server.NewCreateServerCommand(client, *request)

	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	} else {

		d.SetId(resp.Id)
		d.Set("password", resp.Password)
		if resp.OsConfiguration != nil {
			d.Set("root_password", resp.OsConfiguration.RootPassword)
			d.Set("management_ui_url", resp.OsConfiguration.ManagementUiUrl)
		}

		waitResultError := resourceWaitForCreate(resp.Id, &client)
		if waitResultError != nil {
			return waitResultError
		}
	}

	return resourceServerRead(d, m)
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	client := m.(receiver.BMCSDK)
	serverID := d.Id()
	requestCommand := server.NewGetServerCommand(client, serverID)
	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	}

	d.Set("status", resp.Status)
	d.Set("hostname", resp.Hostname)
	d.Set("description", resp.Description)
	d.Set("os", resp.Os)
	d.Set("type", resp.Type)
	d.Set("location", resp.Location)
	d.Set("cpu", resp.Cpu)
	d.Set("cpu_count", resp.CpuCount)
	d.Set("cores_per_cpu", resp.CoresPerCpu)
	d.Set("cpu_frequency_in_ghz", resp.CpuFrequency)
	d.Set("ram", resp.Ram)
	d.Set("storage", resp.Storage)
	d.Set("network_type", resp.NetworkType)
	d.Set("action", "")
	var privateIPs []interface{}
	for _, v := range resp.PrivateIpAddresses {
		privateIPs = append(privateIPs, v)
	}
	d.Set("private_ip_addresses", privateIPs)
	var publicIPs []interface{}
	for _, k := range resp.PublicIpAddresses {
		publicIPs = append(publicIPs, k)
	}
	d.Set("public_ip_addresses", publicIPs)
	d.Set("reservation_id", resp.ReservationId)
	d.Set("pricing_model", resp.PricingModel)

	d.Set("cluster_id", resp.ClusterId)
	if resp.OsConfiguration != nil && resp.OsConfiguration.ManagementAccessAllowedIps != nil {
		var mgmntAccessAllowedIps []interface{}
		for _, k := range *resp.OsConfiguration.ManagementAccessAllowedIps {
			mgmntAccessAllowedIps = append(mgmntAccessAllowedIps, k)
		}
		d.Set("management_access_allowed_ips", mgmntAccessAllowedIps)
	}

	if resp.OsConfiguration != nil && resp.OsConfiguration.Windows != nil && resp.OsConfiguration.Windows.RdpAllowedIps != nil {
		var rdpAllowedIps []interface{}
		for _, k := range *resp.OsConfiguration.Windows.RdpAllowedIps {
			rdpAllowedIps = append(rdpAllowedIps, k)
		}
		d.Set("rdp_allowed_ips", rdpAllowedIps)
	}

	if resp.ProvisionedOn != nil {
		d.Set("provisioned_on", resp.ProvisionedOn.String())
	}

	if resp.Tags != nil && len(*resp.Tags) > 0 {
		tagsRead := *resp.Tags
		var tagsInput = d.Get("tags").([]interface{})
		tags := flattenServerTags(tagsRead, tagsInput)
		if err := d.Set("tags", tags); err != nil {
			return err
		}
	}

	var ncInput = d.Get("network_configuration").([]interface{})
	networkConfiguration := flattenNetworkConfiguration(&resp.NetworkConfiguration, ncInput)

	if err := d.Set("network_configuration", networkConfiguration); err != nil {
		return err
	}

	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	if d.HasChange("action") {
		client := m.(receiver.BMCSDK)
		//var requestCommand helpercommand.Executor
		newStatus := d.Get("action").(string)

		switch newStatus {
		case "powered-on":
			//do power-on request
			serverID := d.Id()
			requestCommand := server.NewPowerOnServerCommand(client, serverID)
			_, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			waitResultError := resourceWaitForPowerON(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}
		case "powered-off":
			//power off request

			serverID := d.Id()

			requestCommand := server.NewPowerOffServerCommand(client, serverID)
			_, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			waitResultError := resourceWaitForPowerOff(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}
		case "reboot":
			//reboot

			serverID := d.Id()

			requestCommand := server.NewRebootServerCommand(client, serverID)
			_, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			waitResultError := resourceWaitForCreate(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}
		case "reset":
			//reset
			request := &bmcapiclient.ServerReset{}
			temp := d.Get("ssh_keys").(*schema.Set).List()
			keys := make([]string, len(temp))
			for i, v := range temp {
				keys[i] = fmt.Sprint(v)
			}
			request.SshKeys = &keys
			request.InstallDefaultSshKeys = d.Get("install_default_ssh_keys").(*bool)

			temp1 := d.Get("ssh_key_ids").(*schema.Set).List()
			keyIds := make([]string, len(temp1))
			for i, v := range temp1 {
				keyIds[i] = fmt.Sprint(v)
			}
			request.SshKeyIds = &keyIds

			dtoOsConfiguration := bmcapiclient.OsConfigurationMap{}
			isWindows := strings.Contains(d.Get("os").(string), "windows")
			isEsxi := strings.Contains(d.Get("os").(string), "esxi")

			if isWindows {
				//log.Printf("Waiting for server windows to be reseted...")
				dtoWindows := bmcapiclient.OsConfigurationWindows{}
				temp2 := d.Get("rdp_allowed_ips").(*schema.Set).List()
				allowedIps := make([]string, len(temp2))
				for i, v := range temp2 {
					allowedIps[i] = fmt.Sprint(v)
				}

				dtoWindows.RdpAllowedIps = &allowedIps
				dtoOsConfiguration.Windows = &dtoWindows
				dtoOsConfiguration.Esxi = nil
				request.OsConfiguration = &dtoOsConfiguration
			}

			if isEsxi {
				//log.Printf("Waiting for server esxi to be reseted...")
				dtoEsxi := bmcapiclient.OsConfigurationMapEsxi{}
				temp3 := d.Get("management_access_allowed_ips").(*schema.Set).List()
				managementAccessAllowedIps := make([]string, len(temp3))
				for i, v := range temp3 {
					managementAccessAllowedIps[i] = fmt.Sprint(v)
				}
				dtoEsxi.ManagementAccessAllowedIps = &managementAccessAllowedIps
				dtoOsConfiguration.Esxi = &dtoEsxi
				dtoOsConfiguration.Windows = nil
				request.OsConfiguration = &dtoOsConfiguration

			}
			requestCommand := server.NewResetServerCommand(client, d.Id(), *request)
			resp, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			d.Set("password", resp.Password)

			if resp.OsConfiguration != nil && resp.OsConfiguration.Esxi != nil {
				d.Set("root_password", resp.OsConfiguration.Esxi.RootPassword)
				d.Set("management_ui_url", resp.OsConfiguration.Esxi.ManagementUiUrl)
			}

			waitResultError := resourceWaitForCreate(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}

		case "shutdown":

			serverID := d.Id()

			requestCommand := server.NewShutDownServerCommand(client, serverID)
			_, err := requestCommand.Execute()
			if err != nil {
				return err
			}
			waitResultError := resourceWaitForPowerOff(d.Id(), &client)
			if waitResultError != nil {
				return waitResultError
			}

		case "default":
			return fmt.Errorf("unsupported action")
		}

	} else if d.HasChange("pricing_model") {
		client := m.(receiver.BMCSDK)
		//var requestCommand command.Executor
		//reserve action
		request := &bmcapiclient.ServerReserve{}
		//request.Id = d.Id()
		request.PricingModel = d.Get("pricing_model").(string)

		requestCommand := server.NewReserveServerCommand(client, d.Id(), *request)
		_, err := requestCommand.Execute()
		if err != nil {
			return err
		}
	} else if d.HasChange("tags") {
		tags := d.Get("tags").([]interface{})
		client := m.(receiver.BMCSDK)
		serverID := d.Id()

		var request []bmcapiclient.TagAssignmentRequest

		if len(tags) > 0 {
			request = make([]bmcapiclient.TagAssignmentRequest, len(tags))

			for i, j := range tags {
				tarObject := bmcapiclient.TagAssignmentRequest{}
				tagsItem := j.(map[string]interface{})

				tagAssign := tagsItem["tag_assignment"].([]interface{})[0]
				tagAssignItem := tagAssign.(map[string]interface{})

				tarObject.Name = tagAssignItem["name"].(string)
				value := tagAssignItem["value"].(string)
				if len(value) > 0 {
					tarObject.Value = &value
				}
				request[i] = tarObject
			}
		}
		requestCommand := server.NewSetServerTagsCommand(client, serverID, request)
		_, err := requestCommand.Execute()
		if err != nil {
			return err
		}

	} else if d.HasChange("network_configuration") {
		client := m.(receiver.BMCSDK)
		serverID := d.Id()
		requestCommand := server.NewGetServerCommand(client, serverID)
		resp, err := requestCommand.Execute()
		if err != nil {
			return err
		}
		netConf := resp.NetworkConfiguration

		var ipBlocks = []bmcapiclient.ServerIpBlock{}
		if netConf.IpBlocksConfiguration != nil {
			ipBlocksConf := *netConf.IpBlocksConfiguration
			if ipBlocksConf.IpBlocks != nil {
				ipBlocks = *ipBlocksConf.IpBlocks
			}
		}
		var privateNetworks = []bmcapiclient.ServerPrivateNetwork{}
		if netConf.PrivateNetworkConfiguration != nil {
			privateNetConf := *netConf.PrivateNetworkConfiguration
			if privateNetConf.PrivateNetworks != nil {
				privateNetworks = *privateNetConf.PrivateNetworks
			}
		}
		var publicNetworks = []bmcapiclient.ServerPublicNetwork{}
		if netConf.PublicNetworkConfiguration != nil {
			publicNetConf := *netConf.PublicNetworkConfiguration
			if publicNetConf.PublicNetworks != nil {
				publicNetworks = *publicNetConf.PublicNetworks
			}
		}
		var ncInput = d.Get("network_configuration").([]interface{})
		var deleteIpBlocks = false

		var ipBlocksInput []interface{}
		var privateNetworksInput []interface{}
		var publicNetworksInput []interface{}

		if len(ncInput) > 0 {
			nci := ncInput[0]
			nciMap := nci.(map[string]interface{})
			ibc := nciMap["ip_blocks_configuration"]
			if ibc != nil && len(ibc.([]interface{})) > 0 {
				ibci := ibc.([]interface{})[0]
				ibcInput := ibci.(map[string]interface{})
				if ibcInput["ip_blocks"] != nil && len(ibcInput["ip_blocks"].([]interface{})) > 0 {
					ipBlocksInput = ibcInput["ip_blocks"].([]interface{})
				}
				if ibcInput["relinquish_ip_blocks"].(bool) {
					deleteIpBlocks = true
				}
			}
			privnc := nciMap["private_network_configuration"]
			if privnc != nil && len(privnc.([]interface{})) > 0 {
				privnci := privnc.([]interface{})[0]
				privncInput := privnci.(map[string]interface{})
				if privncInput["private_networks"] != nil && len(privncInput["private_networks"].([]interface{})) > 0 {
					privateNetworksInput = privncInput["private_networks"].([]interface{})
				}
			}
			pubnc := nciMap["public_network_configuration"]
			if pubnc != nil && len(pubnc.([]interface{})) > 0 {
				pubnci := pubnc.([]interface{})[0]
				pubncInput := pubnci.(map[string]interface{})
				if pubncInput["public_networks"] != nil && len(pubncInput["public_networks"].([]interface{})) > 0 {
					publicNetworksInput = pubncInput["public_networks"].([]interface{})
				}
			}
		}
		var sameIpBlocks []interface{}

		if len(ipBlocks) > 0 && len(ipBlocksInput) > 0 {
			for _, j := range ipBlocks {
				id := j.Id
				for _, l := range ipBlocksInput {
					ipbsInputItem := l.(map[string]interface{})
					if ipbsInputItem["server_ip_block"] != nil && len(ipbsInputItem["server_ip_block"].([]interface{})) > 0 {
						sibInput := ipbsInputItem["server_ip_block"].([]interface{})[0]
						sibInputItem := sibInput.(map[string]interface{})
						idInput := sibInputItem["id"].(string)
						if id == idInput {
							sameIpBlocks = append(sameIpBlocks, id)
						}
					}
				}
			}
		}
		if len(ipBlocks) > len(sameIpBlocks) {
			for _, j := range ipBlocks {
				ipBlockID := j.Id
				var same = false
				for _, l := range sameIpBlocks {
					if ipBlockID == l {
						same = true
					}
				}
				if !same {
					relinquishIpBlock := bmcapiclient.RelinquishIpBlock{}
					relinquishIpBlock.DeleteIpBlocks = &deleteIpBlocks

					requestCommand := server.NewRemoveIpBlockCommand(client, serverID, ipBlockID, relinquishIpBlock)
					_, err := requestCommand.Execute()
					if err != nil {
						return err
					}
					if deleteIpBlocks {
						time.Sleep(time.Minute)
					} else {
						waitResultError := ipBlockWaitForUnassign(ipBlockID, &client)
						if waitResultError != nil {
							return waitResultError
						}
					}
				}
			}
		}
		if len(ipBlocksInput) > len(sameIpBlocks) {
			for _, l := range ipBlocksInput {
				ipbsInputItem := l.(map[string]interface{})
				if ipbsInputItem["server_ip_block"] != nil && len(ipbsInputItem["server_ip_block"].([]interface{})) > 0 {
					sibInput := ipbsInputItem["server_ip_block"].([]interface{})[0]
					sibInputItem := sibInput.(map[string]interface{})
					idInput := sibInputItem["id"].(string)
					vlanId := int32(sibInputItem["vlan_id"].(int))

					var same = false
					for _, l := range sameIpBlocks {
						if idInput == l {
							same = true
						}
					}
					if !same {
						sibObject := bmcapiclient.ServerIpBlock{}
						sibObject.Id = idInput
						sibObject.VlanId = &vlanId
						request := &sibObject
						requestCommand := server.NewAddIpBlockCommand(client, serverID, *request)
						_, err := requestCommand.Execute()
						if err != nil {
							return err
						}
						waitResultError := ipBlockWaitForAssign(idInput, &client)
						if waitResultError != nil {
							return waitResultError
						}
					}
				}
			}
		}
		var samePrivateNetworks []interface{}

		if len(privateNetworks) > 0 && len(privateNetworksInput) > 0 {
			for _, j := range privateNetworks {
				id := j.Id
				for _, l := range privateNetworksInput {
					privateNetworksInputItem := l.(map[string]interface{})
					if privateNetworksInputItem["server_private_network"] != nil && len(privateNetworksInputItem["server_private_network"].([]interface{})) > 0 {
						spnInput := privateNetworksInputItem["server_private_network"].([]interface{})[0]
						spnInputItem := spnInput.(map[string]interface{})
						idInput := spnInputItem["id"].(string)
						if id == idInput {
							samePrivateNetworks = append(samePrivateNetworks, id)
						}
					}
				}
			}
		}
		if len(privateNetworks) > len(samePrivateNetworks) {
			for _, j := range privateNetworks {
				privateNetworkID := j.Id
				var same = false
				for _, l := range samePrivateNetworks {
					if privateNetworkID == l {
						same = true
					}
				}
				if !same {
					requestCommand := server.NewDeleteServerPrivateNetworkCommand(client, serverID, privateNetworkID)
					_, err := requestCommand.Execute()
					if err != nil {
						return err
					}
				}
			}
		}
		if len(privateNetworksInput) > len(samePrivateNetworks) {
			for _, l := range privateNetworksInput {
				privateNetworksInputItem := l.(map[string]interface{})
				if privateNetworksInputItem["server_private_network"] != nil && len(privateNetworksInputItem["server_private_network"].([]interface{})) > 0 {
					spnInput := privateNetworksInputItem["server_private_network"].([]interface{})[0]
					spnInputItem := spnInput.(map[string]interface{})
					idInput := spnInputItem["id"].(string)
					var same = false
					for _, l := range samePrivateNetworks {
						if idInput == l {
							same = true
						}
					}
					if !same {
						spnObject := bmcapiclient.ServerPrivateNetwork{}
						spnObject.Id = idInput

						tempIps := spnInputItem["ips"].(*schema.Set).List()
						netIps := make([]string, len(tempIps))
						for i, v := range tempIps {
							netIps[i] = fmt.Sprint(v)
						}
						if (len(netIps)) > 0 {
							spnObject.Ips = &netIps
						}

						dhcp := spnInputItem["dhcp"].(bool)
						spnObject.Dhcp = &dhcp

						request := &spnObject
						requestCommand := server.NewAddServer2PrivateNetworkCommand(client, serverID, *request)
						_, err := requestCommand.Execute()
						if err != nil {
							return err
						}
					}
				}
			}
		}
		var samePublicNetworks []interface{}

		if len(publicNetworks) > 0 && len(publicNetworksInput) > 0 {
			for _, j := range publicNetworks {
				id := j.Id
				for _, l := range publicNetworksInput {
					publicNetworksInputItem := l.(map[string]interface{})
					if publicNetworksInputItem["server_public_network"] != nil && len(publicNetworksInputItem["server_public_network"].([]interface{})) > 0 {
						spnInput := publicNetworksInputItem["server_public_network"].([]interface{})[0]
						spnInputItem := spnInput.(map[string]interface{})
						idInput := spnInputItem["id"].(string)
						if id == idInput {
							samePublicNetworks = append(samePublicNetworks, id)
						}
					}
				}
			}
		}
		if len(publicNetworks) > len(samePublicNetworks) {
			for _, j := range publicNetworks {
				publicNetworkID := j.Id
				var same = false
				for _, l := range samePublicNetworks {
					if publicNetworkID == l {
						same = true
					}
				}
				if !same {
					requestCommand := server.NewDeleteServerPublicNetworkCommand(client, serverID, publicNetworkID)
					_, err := requestCommand.Execute()
					if err != nil {
						return err
					}
				}
			}
		}
		if len(publicNetworksInput) > len(samePublicNetworks) {
			for _, l := range publicNetworksInput {
				publicNetworksInputItem := l.(map[string]interface{})
				if publicNetworksInputItem["server_public_network"] != nil && len(publicNetworksInputItem["server_public_network"].([]interface{})) > 0 {
					spnInput := publicNetworksInputItem["server_public_network"].([]interface{})[0]
					spnInputItem := spnInput.(map[string]interface{})
					idInput := spnInputItem["id"].(string)
					var same = false
					for _, l := range samePublicNetworks {
						if idInput == l {
							same = true
						}
					}
					if !same {
						spnObject := bmcapiclient.ServerPublicNetwork{}
						spnObject.Id = idInput

						tempIps := spnInputItem["ips"].(*schema.Set).List()
						netIps := make([]string, len(tempIps))
						for i, v := range tempIps {
							netIps[i] = fmt.Sprint(v)
						}
						if (len(netIps)) > 0 {
							spnObject.Ips = netIps
						}
						request := &spnObject

						requestCommand := server.NewAddServer2PublicNetworkCommand(client, serverID, *request)
						_, err := requestCommand.Execute()
						if err != nil {
							return err
						}
					}
				}
			}
		}
	} else {
		return fmt.Errorf("unsupported action")
	}
	return resourceServerRead(d, m)
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(receiver.BMCSDK)
	serverID := d.Id()
	var ncInput = d.Get("network_configuration").([]interface{})
	var ibConfigurationType string
	var relinquishIpBlocks = false
	if len(ncInput) > 0 {
		nci := ncInput[0]
		nciMap := nci.(map[string]interface{})
		ibc := nciMap["ip_blocks_configuration"]
		if ibc != nil && len(ibc.([]interface{})) > 0 {
			ibci := ibc.([]interface{})[0]
			ibcInput := ibci.(map[string]interface{})
			if ibcInput["configuration_type"] != nil && len(ibcInput["configuration_type"].(string)) > 0 {
				ibConfigurationType = ibcInput["configuration_type"].(string)
			}
			relinquishIpBlocks = ibcInput["relinquish_ip_blocks"].(bool)
		}
	}
	var deleteIpBlocks = true
	if ibConfigurationType == "USER_DEFINED" {
		deleteIpBlocks = false
	}
	if relinquishIpBlocks {
		deleteIpBlocks = true
	}
	relinquishIpBlock := bmcapiclient.RelinquishIpBlock{}
	relinquishIpBlock.DeleteIpBlocks = &deleteIpBlocks
	b, _ := json.MarshalIndent(relinquishIpBlock, "", "  ")
	log.Printf("relinquishIpBlock object is" + string(b))

	requestCommand := server.NewDeprovisionServerCommand(client, serverID, relinquishIpBlock)
	_, err := requestCommand.Execute()
	if err != nil {
		return err
	}
	return nil
}

func resourceWaitForCreate(id string, client *receiver.BMCSDK) error {
	log.Printf("Waiting for server %s to be created...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating", "resetting", "rebooting"},
		Target:     []string{"powered-on", "powered-off"},
		Refresh:    refreshForCreate(client, id),
		Timeout:    pnapRetryTimeout,
		Delay:      pnapRetryDelay,
		MinTimeout: pnapRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for server (%s) to switch to target state: %v", id, err)
	}

	return nil
}

func resourceWaitForPowerON(id string, client *receiver.BMCSDK) error {
	log.Printf("Waiting for server %s to power on...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"powered-off"},
		Target:     []string{"powered-on"},
		Refresh:    refreshForCreate(client, id),
		Timeout:    pnapRetryTimeout,
		Delay:      pnapRetryDelay,
		MinTimeout: pnapRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for server (%s) to power on: %v", id, err)
	}

	return nil
}

func resourceWaitForPowerOff(id string, client *receiver.BMCSDK) error {
	log.Printf("Waiting for server %s to power off...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"powered-on"},
		Target:     []string{"powered-off"},
		Refresh:    refreshForCreate(client, id),
		Timeout:    pnapRetryTimeout,
		Delay:      pnapRetryDelay,
		MinTimeout: pnapRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for server (%s) to power off: %v", id, err)
	}

	return nil
}

func refreshForCreate(client *receiver.BMCSDK, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		requestCommand := server.NewGetServerCommand(*client, id)

		resp, err := requestCommand.Execute()
		if err != nil {
			return 0, "", err
		} else {
			return 0, resp.Status, nil
		}
	}
}

func ipBlockWaitForUnassign(id string, client *receiver.BMCSDK) error {
	log.Printf("Waiting for ip block %s to be unassigned...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"assigned", "unassigning"},
		Target:     []string{"unassigned"},
		Refresh:    refreshForIpBlockStatus(client, id),
		Timeout:    pnapRetryTimeout,
		Delay:      pnapRetryDelay,
		MinTimeout: pnapRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for ip block (%s) to be unassigned: %v", id, err)
	}

	return nil
}

func ipBlockWaitForAssign(id string, client *receiver.BMCSDK) error {
	log.Printf("Waiting for ip block %s to be unassigned...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"unassigned", "assigning"},
		Target:     []string{"assigned"},
		Refresh:    refreshForIpBlockStatus(client, id),
		Timeout:    pnapRetryTimeout,
		Delay:      pnapRetryDelay,
		MinTimeout: pnapRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for ip block (%s) to be assigned: %v", id, err)
	}

	return nil
}

func refreshForIpBlockStatus(client *receiver.BMCSDK, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		requestCommand := ipblock.NewGetIpBlockCommand(*client, id)

		resp, err := requestCommand.Execute()
		if err != nil {
			return 0, "", err
		} else {
			return 0, resp.Status, nil
		}
	}
}

func flattenNetworkConfiguration(netConf *bmcapiclient.NetworkConfiguration, ncInput []interface{}) []interface{} {
	if netConf != nil {
		if len(ncInput) == 0 {
			ncInput = make([]interface{}, 1)
			n := make(map[string]interface{})
			ncInput[0] = n
		}
		nci := ncInput[0]
		nciMap := nci.(map[string]interface{})

		if netConf.GatewayAddress != nil {
			nciMap["gateway_address"] = *netConf.GatewayAddress
		}
		prNetConfA := netConf.PrivateNetworkConfiguration
		if prNetConfA != nil && *prNetConfA.PrivateNetworks != nil && len(*prNetConfA.PrivateNetworks) > 0 {
			prNetConf := *netConf.PrivateNetworkConfiguration
			//pnc := make([]interface{}, 1)
			var pnc []interface{}
			if (nciMap["private_network_configuration"]) != nil && len(nciMap["private_network_configuration"].([]interface{})) > 0 {
				pnc = nciMap["private_network_configuration"].([]interface{})
			} else {
				pnc = make([]interface{}, 1)
			}
			//pncItem := make(map[string]interface{})
			var pncItem map[string]interface{}
			if len(pnc) > 0 && pnc[0] != nil {
				pncItem = pnc[0].(map[string]interface{})
			} else {
				pncItem = make(map[string]interface{})
			}
			if prNetConf.GatewayAddress != nil {
				pncItem["gateway_adress"] = *prNetConf.GatewayAddress
			}
			if prNetConf.ConfigurationType != nil && len(*prNetConf.ConfigurationType) > 0 {
				pncItem["configuration_type"] = *prNetConf.ConfigurationType
			}
			if prNetConf.PrivateNetworks != nil {
				prNet := *prNetConf.PrivateNetworks
				var pn []interface{}
				if pncItem["private_networks"] != nil {
					pn = pncItem["private_networks"].([]interface{})
				} else {
					pn = make([]interface{}, len(prNet))
				}
				privateNetworks := sortPrivateNetworks(prNet, pn)
				pncItem["private_networks"] = privateNetworks
			}
			pnc[0] = pncItem
			nciMap["private_network_configuration"] = pnc
		} else {
			nciMap["private_network_configuration"] = make([]interface{}, 0)
		}
		ipBlocksConfA := netConf.IpBlocksConfiguration
		if ipBlocksConfA != nil && *ipBlocksConfA.IpBlocks != nil && len(*ipBlocksConfA.IpBlocks) > 0 {
			ipBlocksConf := *netConf.IpBlocksConfiguration
			var ibc []interface{}
			if (nciMap["ip_blocks_configuration"]) != nil && len(nciMap["ip_blocks_configuration"].([]interface{})) > 0 {
				ibc = nciMap["ip_blocks_configuration"].([]interface{})
			} else {
				ibc = make([]interface{}, 1)
			}
			var ibcItem map[string]interface{}
			if len(ibc) > 0 && ibc[0] != nil {
				ibcItem = ibc[0].(map[string]interface{})
			} else {
				ibcItem = make(map[string]interface{})
			}
			if ipBlocksConf.ConfigurationType != nil && len(*ipBlocksConf.ConfigurationType) > 0 {
				ibcItem["configuration_type"] = *ipBlocksConf.ConfigurationType
			}
			if ipBlocksConf.IpBlocks != nil {
				ipBloc := *ipBlocksConf.IpBlocks
				var ib []interface{}
				if ibcItem["ip_blocks"] != nil {
					ib = ibcItem["ip_blocks"].([]interface{})
				} else {
					ib = make([]interface{}, len(ipBloc))
				}
				ipBlocks := sortIpBlocks(ipBloc, ib)
				ibcItem["ip_blocks"] = ipBlocks
			}
			ibc[0] = ibcItem
			nciMap["ip_blocks_configuration"] = ibc
		} else {
			nciMap["ip_blocks_configuration"] = make([]interface{}, 0)
		}
		puNetConfA := netConf.PublicNetworkConfiguration
		if puNetConfA != nil && *puNetConfA.PublicNetworks != nil && len(*puNetConfA.PublicNetworks) > 0 {
			puNetConf := *netConf.PublicNetworkConfiguration
			var pnc []interface{}
			if (nciMap["public_network_configuration"]) != nil && len(nciMap["public_network_configuration"].([]interface{})) > 0 {
				pnc = nciMap["public_network_configuration"].([]interface{})
			} else {
				pnc = make([]interface{}, 1)
			}
			var pncItem map[string]interface{}
			if len(pnc) > 0 && pnc[0] != nil {
				pncItem = pnc[0].(map[string]interface{})
			} else {
				pncItem = make(map[string]interface{})
			}
			if puNetConf.PublicNetworks != nil {
				puNet := *puNetConf.PublicNetworks
				var pn []interface{}
				if pncItem["public_networks"] != nil {
					pn = pncItem["public_networks"].([]interface{})
				} else {
					pn = make([]interface{}, len(puNet))
				}
				publicNetworks := sortPublicNetworks(puNet, pn)
				pncItem["public_networks"] = publicNetworks
			}
			pnc[0] = pncItem
			nciMap["public_network_configuration"] = pnc
		} else {
			nciMap["public_network_configuration"] = make([]interface{}, 0)
		}
		return ncInput
	}
	return make([]interface{}, 0)
}

func sortPrivateNetworks(prNet []bmcapiclient.ServerPrivateNetwork, prNetInput []interface{}) []interface{} {
	if len(prNet) > 0 {
		var pn []interface{}
		if len(prNetInput) == 0 || prNetInput[0] == nil {
			for _, j := range prNet {
				pnItem := make(map[string]interface{})
				spn := make([]interface{}, 1)
				spnItem := make(map[string]interface{})

				spnItem["id"] = j.Id
				if j.Ips != nil {
					ips := make([]interface{}, len(*j.Ips))
					for k, l := range *j.Ips {
						ips[k] = l
					}
					spnItem["ips"] = ips
				}
				if j.Dhcp != nil {
					spnItem["dhcp"] = *j.Dhcp
				}
				if j.StatusDescription != nil {
					spnItem["status_description"] = *j.StatusDescription
				}
				spn[0] = spnItem
				pnItem["server_private_network"] = spn
				pn = append(pn, pnItem)
			}
		} else if len(prNetInput) > 0 {
			for i := range prNetInput {
				for _, j := range prNet {
					if prNetInput[i].(map[string]interface{})["server_private_network"].([]interface{})[0].(map[string]interface{})["id"] == j.Id {
						pnItem := make(map[string]interface{})
						spn := make([]interface{}, 1)
						spnItem := make(map[string]interface{})

						spnItem["id"] = j.Id
						if j.Ips != nil {
							ips := make([]interface{}, len(*j.Ips))
							for k, l := range *j.Ips {
								ips[k] = l
							}
							spnItem["ips"] = ips
						}
						if j.Dhcp != nil {
							spnItem["dhcp"] = *j.Dhcp
						}
						if j.StatusDescription != nil {
							spnItem["status_description"] = *j.StatusDescription
						}
						spn[0] = spnItem
						pnItem["server_private_network"] = spn
						pn = append(pn, pnItem)
					}
				}
			}
			for _, p := range prNet {
				var newPrivateNetwork = true
				for r := range pn {
					if p.Id == pn[r].(map[string]interface{})["server_private_network"].([]interface{})[0].(map[string]interface{})["id"] {
						newPrivateNetwork = false
					}
				}
				if newPrivateNetwork {
					pnItem := make(map[string]interface{})
					spn := make([]interface{}, 1)
					spnItem := make(map[string]interface{})

					spnItem["id"] = p.Id
					if p.Ips != nil {
						ips := make([]interface{}, len(*p.Ips))
						for k, l := range *p.Ips {
							ips[k] = l
						}
						spnItem["ips"] = ips
					}
					if p.Dhcp != nil {
						spnItem["dhcp"] = *p.Dhcp
					}
					if p.StatusDescription != nil {
						spnItem["status_description"] = *p.StatusDescription
					}
					spn[0] = spnItem
					pnItem["server_private_network"] = spn
					pn = append(pn, pnItem)
				}
			}
		}
		return pn
	}
	return make([]interface{}, 0)
}

func sortIpBlocks(ipBlocks []bmcapiclient.ServerIpBlock, ibInput []interface{}) []interface{} {
	if len(ipBlocks) > 0 {
		var ipb []interface{}
		if len(ibInput) == 0 || ibInput[0] == nil {
			for _, j := range ipBlocks {
				ipbItem := make(map[string]interface{})
				sipb := make([]interface{}, 1)
				sipbItem := make(map[string]interface{})

				sipbItem["id"] = j.Id
				if j.VlanId != nil {
					sipbItem["vlan_id"] = int(*j.VlanId)
				}
				sipb[0] = sipbItem
				ipbItem["server_ip_block"] = sipb
				ipb = append(ipb, ipbItem)
			}
		} else if len(ibInput) > 0 {
			for i := range ibInput {
				for _, l := range ipBlocks {
					if ibInput[i].(map[string]interface{})["server_ip_block"].([]interface{})[0].(map[string]interface{})["id"] == l.Id {
						ipbItem := make(map[string]interface{})
						sipb := make([]interface{}, 1)
						sipbItem := make(map[string]interface{})

						sipbItem["id"] = l.Id
						if l.VlanId != nil {
							sipbItem["vlan_id"] = int(*l.VlanId)
						}
						sipb[0] = sipbItem
						ipbItem["server_ip_block"] = sipb
						ipb = append(ipb, ipbItem)
					}
				}
			}
			for _, p := range ipBlocks {
				var newIpBlock = true
				for r := range ipb {
					if p.Id == ipb[r].(map[string]interface{})["server_ip_block"].([]interface{})[0].(map[string]interface{})["id"] {
						newIpBlock = false
					}
				}
				if newIpBlock {
					ipbItem := make(map[string]interface{})
					sipb := make([]interface{}, 1)
					sipbItem := make(map[string]interface{})

					sipbItem["id"] = p.Id
					if p.VlanId != nil {
						sipbItem["vlan_id"] = int(*p.VlanId)
					}
					sipb[0] = sipbItem
					ipbItem["server_ip_block"] = sipb
					ipb = append(ipb, ipbItem)
				}
			}
		}
		return ipb
	}
	return make([]interface{}, 0)
}

func sortPublicNetworks(puNet []bmcapiclient.ServerPublicNetwork, puNetInput []interface{}) []interface{} {
	if len(puNet) > 0 {
		var pn []interface{}
		if len(puNetInput) == 0 || puNetInput[0] == nil {
			for _, j := range puNet {
				pnItem := make(map[string]interface{})
				spn := make([]interface{}, 1)
				spnItem := make(map[string]interface{})

				spnItem["id"] = j.Id
				if j.Ips != nil {
					ips := make([]interface{}, len(j.Ips))
					for k, l := range j.Ips {
						ips[k] = l
					}
					spnItem["ips"] = ips
				}
				if j.StatusDescription != nil {
					spnItem["status_description"] = *j.StatusDescription
				}
				spn[0] = spnItem
				pnItem["server_public_network"] = spn
				pn = append(pn, pnItem)
			}
		} else if len(puNetInput) > 0 {
			for i := range puNetInput {
				for _, j := range puNet {
					if puNetInput[i].(map[string]interface{})["server_public_network"].([]interface{})[0].(map[string]interface{})["id"] == j.Id {
						pnItem := make(map[string]interface{})
						spn := make([]interface{}, 1)
						spnItem := make(map[string]interface{})

						spnItem["id"] = j.Id
						if j.Ips != nil {
							ips := make([]interface{}, len(j.Ips))
							for k, l := range j.Ips {
								ips[k] = l
							}
							spnItem["ips"] = ips
						}
						if j.StatusDescription != nil {
							spnItem["status_description"] = *j.StatusDescription
						}
						spn[0] = spnItem
						pnItem["server_public_network"] = spn
						pn = append(pn, pnItem)
					}
				}
			}
			for _, p := range puNet {
				var newPublicNetwork = true
				for r := range pn {
					if p.Id == pn[r].(map[string]interface{})["server_public_network"].([]interface{})[0].(map[string]interface{})["id"] {
						newPublicNetwork = false
					}
				}
				if newPublicNetwork {
					pnItem := make(map[string]interface{})
					spn := make([]interface{}, 1)
					spnItem := make(map[string]interface{})

					spnItem["id"] = p.Id
					if p.Ips != nil {
						ips := make([]interface{}, len(p.Ips))
						for k, l := range p.Ips {
							ips[k] = l
						}
						spnItem["ips"] = ips
					}
					if p.StatusDescription != nil {
						spnItem["status_description"] = *p.StatusDescription
					}
					spn[0] = spnItem
					pnItem["server_public_network"] = spn
					pn = append(pn, pnItem)
				}
			}
		}
		return pn
	}
	return make([]interface{}, 0)
}

func flattenServerTags(tagsRead []bmcapiclient.TagAssignment, tagsInput []interface{}) []interface{} {
	if len(tagsRead) > 0 {
		var tags []interface{}
		if len(tagsInput) == 0 || tagsInput[0] == nil {
			for _, j := range tagsRead {
				tagsItem := make(map[string]interface{})
				ta := make([]interface{}, 1)
				taItem := make(map[string]interface{})

				taItem["id"] = j.Id
				taItem["name"] = j.Name
				if j.Value != nil {
					taItem["value"] = *j.Value
				}
				taItem["is_billing_tag"] = j.IsBillingTag
				if j.CreatedBy != nil {
					taItem["created_by"] = *j.CreatedBy
				}
				ta[0] = taItem
				tagsItem["tag_assignment"] = ta
				tags = append(tags, tagsItem)
			}
		} else if len(tagsInput) > 0 {
			for i := range tagsInput {
				for _, j := range tagsRead {
					if tagsInput[i].(map[string]interface{})["tag_assignment"].([]interface{})[0].(map[string]interface{})["name"] == j.Name {
						tagsItem := make(map[string]interface{})
						ta := make([]interface{}, 1)
						taItem := make(map[string]interface{})

						taItem["id"] = j.Id
						taItem["name"] = j.Name
						if j.Value != nil {
							taItem["value"] = *j.Value
						}
						taItem["is_billing_tag"] = j.IsBillingTag
						if j.CreatedBy != nil {
							taItem["created_by"] = *j.CreatedBy
						}
						ta[0] = taItem
						tagsItem["tag_assignment"] = ta
						tags = append(tags, tagsItem)
					}
				}
			}
			for _, p := range tagsRead {
				var newTag = true
				for r := range tags {
					if p.Name == tags[r].(map[string]interface{})["tag_assignment"].([]interface{})[0].(map[string]interface{})["name"] {
						newTag = false
					}
				}
				if newTag {
					tagsItem := make(map[string]interface{})
					ta := make([]interface{}, 1)
					taItem := make(map[string]interface{})

					taItem["id"] = p.Id
					taItem["name"] = p.Name
					if p.Value != nil {
						taItem["value"] = *p.Value
					}
					taItem["is_billing_tag"] = p.IsBillingTag
					if p.CreatedBy != nil {
						taItem["created_by"] = *p.CreatedBy
					}
					ta[0] = taItem
					tagsItem["tag_assignment"] = ta
					tags = append(tags, tagsItem)
				}
			}
		}
		return tags
	}
	return make([]interface{}, 0)
}
