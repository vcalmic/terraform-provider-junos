package junos

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type securityOptions struct {
	ikeTraceoptions []map[string]interface{}
	utmOptions      []map[string]interface{}
}

func resourceSecurity() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecurityCreate,
		ReadContext:   resourceSecurityRead,
		UpdateContext: resourceSecurityUpdate,
		DeleteContext: resourceSecurityDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSecurityImport,
		},
		Schema: map[string]*schema.Schema{
			"ike_traceoptions": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"files": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(2, 1000),
									},
									"match": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"size": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(10240, 1073741824),
									},
									"no_world_readable": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"world_readable": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"flag": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"no_remote_trace": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"rate_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      -1,
							ValidateFunc: validation.IntBetween(0, 4294967295),
						},
					},
				},
			},
			"utm": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"feature_profile_web_filtering_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"juniper-enhanced", "juniper-local", "web-filtering-none", "websense-redirect",
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourceSecurityCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sess := m.(*Session)
	jnprSess, err := sess.startNewSession()
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(jnprSess)
	if !checkCompatibilitySecurity(jnprSess) {
		return diag.FromErr(fmt.Errorf("security not compatible with Junos device %s", jnprSess.Platform[0].Model))
	}
	sess.configLock(jnprSess)

	err = setSecurity(d, m, jnprSess)
	if err != nil {
		sess.configClear(jnprSess)

		return diag.FromErr(err)
	}
	err = sess.commitConf("create resource junos_security", jnprSess)
	if err != nil {
		sess.configClear(jnprSess)

		return diag.FromErr(err)
	}

	d.SetId("security")

	return resourceSecurityRead(ctx, d, m)
}
func resourceSecurityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sess := m.(*Session)
	mutex.Lock()
	jnprSess, err := sess.startNewSession()
	if err != nil {
		mutex.Unlock()

		return diag.FromErr(err)
	}
	defer sess.closeSession(jnprSess)
	securityOptions, err := readSecurity(m, jnprSess)
	mutex.Unlock()
	if err != nil {
		return diag.FromErr(err)
	}
	fillSecurity(d, securityOptions)

	return nil
}
func resourceSecurityUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	sess := m.(*Session)
	jnprSess, err := sess.startNewSession()
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(jnprSess)
	sess.configLock(jnprSess)
	err = delSecurity(m, jnprSess)
	if err != nil {
		sess.configClear(jnprSess)

		return diag.FromErr(err)
	}
	err = setSecurity(d, m, jnprSess)
	if err != nil {
		sess.configClear(jnprSess)

		return diag.FromErr(err)
	}
	err = sess.commitConf("update resource junos_security", jnprSess)
	if err != nil {
		sess.configClear(jnprSess)

		return diag.FromErr(err)
	}
	d.Partial(false)

	return resourceSecurityRead(ctx, d, m)
}
func resourceSecurityDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
func resourceSecurityImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	sess := m.(*Session)
	jnprSess, err := sess.startNewSession()
	if err != nil {
		return nil, err
	}
	defer sess.closeSession(jnprSess)
	result := make([]*schema.ResourceData, 1)
	securityOptions, err := readSecurity(m, jnprSess)
	if err != nil {
		return nil, err
	}
	fillSecurity(d, securityOptions)
	d.SetId("security")
	result[0] = d

	return result, nil
}

func setSecurity(d *schema.ResourceData, m interface{}, jnprSess *NetconfObject) error {
	sess := m.(*Session)

	setPrefix := "set "
	configSet := make([]string, 0)

	for _, ikeTrace := range d.Get("ike_traceoptions").([]interface{}) {
		if ikeTrace != nil {
			ikeTraceM := ikeTrace.(map[string]interface{})
			for _, ikeTraceFile := range ikeTraceM["file"].([]interface{}) {
				if ikeTraceFile != nil {
					ikeTraceFileM := ikeTraceFile.(map[string]interface{})
					if ikeTraceFileM["name"].(string) != "" {
						configSet = append(configSet, setPrefix+"security ike traceoptions file "+
							ikeTraceFileM["name"].(string))
					}
					if ikeTraceFileM["files"].(int) > 0 {
						configSet = append(configSet, setPrefix+"security ike traceoptions file files "+
							strconv.Itoa(ikeTraceFileM["files"].(int)))
					}
					if ikeTraceFileM["match"].(string) != "" {
						configSet = append(configSet, setPrefix+"security ike traceoptions file match \""+
							ikeTraceFileM["match"].(string)+"\"")
					}
					if ikeTraceFileM["size"].(int) > 0 {
						configSet = append(configSet, setPrefix+"security ike traceoptions file size "+
							strconv.Itoa(ikeTraceFileM["size"].(int)))
					}
					if ikeTraceFileM["world_readable"].(bool) && ikeTraceFileM["no_world_readable"].(bool) {
						return fmt.Errorf("conflict between 'world_readable' and 'no_world_readable' for ike_traceoptions file")
					}
					if ikeTraceFileM["world_readable"].(bool) {
						configSet = append(configSet, setPrefix+"security ike traceoptions file world-readable")
					}
					if ikeTraceFileM["no_world_readable"].(bool) {
						configSet = append(configSet, setPrefix+"security ike traceoptions file no-world-readable")
					}
				}
			}
			for _, ikeTraceFlag := range ikeTraceM["flag"].([]interface{}) {
				configSet = append(configSet, setPrefix+"security ike traceoptions flag "+ikeTraceFlag.(string))
			}
			if ikeTraceM["no_remote_trace"].(bool) {
				configSet = append(configSet, setPrefix+"security ike traceoptions no-remote-trace")
			}
			if ikeTraceM["rate_limit"].(int) > -1 {
				configSet = append(configSet, setPrefix+"security ike traceoptions rate-limit "+
					strconv.Itoa(ikeTraceM["rate_limit"].(int)))
			}
		}
	}
	for _, utm := range d.Get("utm").([]interface{}) {
		if utm != nil {
			utmM := utm.(map[string]interface{})
			if utmM["feature_profile_web_filtering_type"].(string) != "" {
				configSet = append(configSet, setPrefix+"security utm feature-profile web-filtering type "+
					utmM["feature_profile_web_filtering_type"].(string))
			}
		}
	}
	err := sess.configSet(configSet, jnprSess)
	if err != nil {
		return err
	}

	return nil
}

func delSecurity(m interface{}, jnprSess *NetconfObject) error {
	listLineToDelete := []string{
		"ike traceoptions",
		"utm feature-profile web-filtering type",
	}
	sess := m.(*Session)
	configSet := make([]string, 0)
	delPrefix := "delete security "
	for _, line := range listLineToDelete {
		configSet = append(configSet,
			delPrefix+line)
	}
	err := sess.configSet(configSet, jnprSess)
	if err != nil {
		return err
	}

	return nil
}
func readSecurity(m interface{}, jnprSess *NetconfObject) (securityOptions, error) {
	sess := m.(*Session)
	var confRead securityOptions

	securityConfig, err := sess.command("show configuration security"+
		" | display set relative", jnprSess)
	if err != nil {
		return confRead, err
	}
	if securityConfig != emptyWord {
		for _, item := range strings.Split(securityConfig, "\n") {
			if strings.Contains(item, "<configuration-output>") {
				continue
			}
			if strings.Contains(item, "</configuration-output>") {
				break
			}
			itemTrim := strings.TrimPrefix(item, setLineStart)
			switch {
			case strings.HasPrefix(itemTrim, "ike traceoptions"):
				ikeTraceoptions := map[string]interface{}{
					"file":            make([]map[string]interface{}, 0),
					"flag":            make([]string, 0),
					"no_remote_trace": false,
					"rate_limit":      -1,
				}
				if len(confRead.ikeTraceoptions) > 0 {
					for k, v := range confRead.ikeTraceoptions[0] {
						ikeTraceoptions[k] = v
					}
				}
				switch {
				case strings.HasPrefix(itemTrim, "ike traceoptions file"):
					ikeTraceOptionsFile := map[string]interface{}{
						"name":              "",
						"files":             0,
						"match":             "",
						"size":              0,
						"world_readable":    false,
						"no_world_readable": false,
					}
					if len(ikeTraceoptions["file"].([]map[string]interface{})) > 0 {
						for k, v := range ikeTraceoptions["file"].([]map[string]interface{})[0] {
							ikeTraceOptionsFile[k] = v
						}
					}
					switch {
					case strings.HasPrefix(itemTrim, "ike traceoptions file files"):
						var err error
						ikeTraceOptionsFile["files"], err = strconv.Atoi(
							strings.TrimPrefix(itemTrim, "ike traceoptions file files "))
						if err != nil {
							return confRead, err
						}
						ikeTraceoptions["file"] = []map[string]interface{}{ikeTraceOptionsFile}
					case strings.HasPrefix(itemTrim, "ike traceoptions file match"):
						ikeTraceOptionsFile["match"] = strings.Trim(
							strings.TrimPrefix(itemTrim, "ike traceoptions file match "), "\"")
						ikeTraceoptions["file"] = []map[string]interface{}{ikeTraceOptionsFile}
					case strings.HasPrefix(itemTrim, "ike traceoptions file size"):
						var err error
						ikeTraceOptionsFile["size"], err = strconv.Atoi(
							strings.TrimPrefix(itemTrim, "ike traceoptions file size "))
						if err != nil {
							return confRead, err
						}
						ikeTraceoptions["file"] = []map[string]interface{}{ikeTraceOptionsFile}
					case strings.HasPrefix(itemTrim, "ike traceoptions file world-readable"):
						ikeTraceOptionsFile["world_readable"] = true
						ikeTraceoptions["file"] = []map[string]interface{}{ikeTraceOptionsFile}
					case strings.HasPrefix(itemTrim, "ike traceoptions file no-world-readable"):
						ikeTraceOptionsFile["no_world_readable"] = true
						ikeTraceoptions["file"] = []map[string]interface{}{ikeTraceOptionsFile}
					case strings.HasPrefix(itemTrim, "ike traceoptions file "):
						ikeTraceOptionsFile["name"] = strings.Trim(
							strings.TrimPrefix(itemTrim, "ike traceoptions file "), "\"")
						ikeTraceoptions["file"] = []map[string]interface{}{ikeTraceOptionsFile}
					}
				case strings.HasPrefix(itemTrim, "ike traceoptions flag"):
					ikeTraceoptions["flag"] = append(ikeTraceoptions["flag"].([]string),
						strings.TrimPrefix(itemTrim, "ike traceoptions flag "))
				case strings.HasPrefix(itemTrim, "ike traceoptions no-remote-trace"):
					ikeTraceoptions["no_remote_trace"] = true
				case strings.HasPrefix(itemTrim, "ike traceoptions rate-limit"):
					var err error
					ikeTraceoptions["rate_limit"], err = strconv.Atoi(
						strings.TrimPrefix(itemTrim, "ike traceoptions rate-limit "))
					if err != nil {
						return confRead, err
					}
				}
				confRead.ikeTraceoptions = []map[string]interface{}{ikeTraceoptions}
			case strings.HasPrefix(itemTrim, "utm "):
				utmOptions := map[string]interface{}{
					"feature_profile_web_filtering_type": "",
				}
				if strings.HasPrefix(itemTrim, "utm feature-profile web-filtering type ") {
					utmOptions["feature_profile_web_filtering_type"] = strings.TrimPrefix(itemTrim,
						"utm feature-profile web-filtering type ")
				}
				confRead.utmOptions = []map[string]interface{}{utmOptions}
			}
		}
	}

	return confRead, nil
}

func fillSecurity(d *schema.ResourceData, securityOptions securityOptions) {
	tfErr := d.Set("ike_traceoptions", securityOptions.ikeTraceoptions)
	if tfErr != nil {
		panic(tfErr)
	}
	tfErr = d.Set("utm", securityOptions.utmOptions)
	if tfErr != nil {
		panic(tfErr)
	}
}
