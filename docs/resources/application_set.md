---
page_title: "Junos: junos_application_set"
---

# junos_application_set

Provides a set of applications resource.

## Example Usage

```hcl
# Add a set of applications
resource junos_application_set "ssh_telnet" {
  name         = "ssh_telnet"
  applications = ["junos-ssh", "junos-telnet"]
}
```

## Argument Reference

The following arguments are supported:

- **name** (Required, String, Forces new resource)  
  Name of application set.
- **applications** (Optional, List of String)  
  List of application names.

## Attributes Reference

The following attributes are exported:

- **id** (String)  
  An identifier for the resource with format `<name>`.

## Import

Junos application set can be imported using an id made up of `<name>`, e.g.

```shell
$ terraform import junos_application_set.ssh_telnet ssh_telnet
```
