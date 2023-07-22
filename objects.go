/*
Copyright 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nftables

import (
	"fmt"
	"io"
)

// Object implementation for Table
func (table *Table) validate(verb verb) error {
	switch verb {
	case addVerb, flushVerb:
		if table.Handle != nil {
			return fmt.Errorf("cannot specify Handle in %s operation", verb)
		}
	case deleteVerb:
		// Handle can be nil or non-nil
	default:
		return fmt.Errorf("%s is not implemented for tables", verb)
	}

	return nil
}

func (table *Table) writeOperation(verb verb, family Family, tableName string, writer io.Writer) {
	// Special case for delete-by-handle
	if verb == deleteVerb && table.Handle != nil {
		fmt.Fprintf(writer, "delete table %s handle %d", family, *table.Handle)
		return
	}

	// All other cases refer to the table by name
	fmt.Fprintf(writer, "%s table %s %s", verb, family, tableName)
	if verb == addVerb && table.Comment != nil {
		fmt.Fprintf(writer, " { comment %q ; }", *table.Comment)
	}
	fmt.Fprintf(writer, "\n")
}

// Object implementation for Chain
func (chain *Chain) validate(verb verb) error {
	if chain.Name == "" {
		return fmt.Errorf("no name specified for chain")
	}

	if chain.Hook == nil && (chain.Type != nil || chain.Priority != nil) {
		return fmt.Errorf("regular chain %q must not specify Type or Priority", chain.Name)
	} else if chain.Hook != nil && (chain.Type == nil || chain.Priority == nil) {
		return fmt.Errorf("base chain %q must specify Type and Priority", chain.Name)
	}

	switch verb {
	case addVerb, flushVerb:
		if chain.Handle != nil {
			return fmt.Errorf("cannot specify Handle in %s operation", verb)
		}
	case deleteVerb:
		// Handle can be nil or non-nil
	default:
		return fmt.Errorf("%s is not implemented for chains", verb)
	}

	return nil
}

func (chain *Chain) writeOperation(verb verb, family Family, table string, writer io.Writer) {
	// Special case for delete-by-handle
	if verb == deleteVerb && chain.Handle != nil {
		fmt.Fprintf(writer, "delete chain %s %s handle %d", family, table, *chain.Handle)
		return
	}

	fmt.Fprintf(writer, "%s chain %s %s %s", verb, family, table, chain.Name)
	if verb == addVerb && (chain.Type != nil || chain.Comment != nil) {
		fmt.Fprintf(writer, " {")

		if chain.Type != nil {
			fmt.Fprintf(writer, " type %s hook %s priority %s ;", *chain.Type, *chain.Hook, *chain.Priority)
		}
		if chain.Comment != nil {
			fmt.Fprintf(writer, " comment %q ;", *chain.Comment)
		}

		fmt.Fprintf(writer, " }")
	}

	fmt.Fprintf(writer, "\n")
}

// Object implementation for Rule
func (rule *Rule) validate(verb verb) error {
	if rule.Chain == "" {
		return fmt.Errorf("no chain name specified for rule")
	}

	if rule.Index != nil && rule.Handle != nil {
		return fmt.Errorf("cannot specify both Index and Handle")
	}

	if (verb == deleteVerb || verb == replaceVerb) && rule.Handle == nil {
		return fmt.Errorf("must specify Handle with %s", verb)
	}

	return nil
}

func (rule *Rule) writeOperation(verb verb, family Family, table string, writer io.Writer) {
	fmt.Fprintf(writer, "%s rule %s %s %s", verb, family, table, rule.Chain)
	if rule.Index != nil {
		fmt.Fprintf(writer, " index %d", *rule.Index)
	} else if rule.Handle != nil {
		fmt.Fprintf(writer, " handle %d", *rule.Handle)
	}

	switch verb {
	case addVerb, insertVerb, replaceVerb:
		fmt.Fprintf(writer, " %s", rule.Rule)

		if rule.Comment != nil {
			fmt.Fprintf(writer, " comment %q", *rule.Comment)
		}
	}

	fmt.Fprintf(writer, "\n")
}

// Object implementation for Set
func (set *Set) validate(verb verb) error {
	if set.Name == "" {
		return fmt.Errorf("no name specified for set")
	}

	switch verb {
	case addVerb:
		if (set.Type == "" && set.TypeOf == "") || (set.Type != "" && set.TypeOf != "") {
			return fmt.Errorf("set must specify either Type or TypeOf")
		}
		fallthrough
	case flushVerb:
		if set.Handle != nil {
			return fmt.Errorf("cannot specify Handle in %s operation", verb)
		}
	case deleteVerb:
		// Handle can be nil or non-nil
	default:
		return fmt.Errorf("%s is not implemented for sets", verb)
	}

	return nil
}

func (set *Set) writeOperation(verb verb, family Family, table string, writer io.Writer) {
	// Special case for delete-by-handle
	if verb == deleteVerb && set.Handle != nil {
		fmt.Fprintf(writer, "delete set %s %s handle %d", family, table, *set.Handle)
		return
	}

	fmt.Fprintf(writer, "%s set %s %s %s", verb, family, table, set.Name)
	if verb == addVerb {
		fmt.Fprintf(writer, " {")

		if set.Type != "" {
			fmt.Fprintf(writer, " type %s ;", set.Type)
		} else {
			fmt.Fprintf(writer, " typeof %s ;", set.TypeOf)
		}

		if len(set.Flags) != 0 {
			fmt.Fprintf(writer, " flags ")
			for i := range set.Flags {
				if i > 0 {
					fmt.Fprintf(writer, ",")
				}
				fmt.Fprintf(writer, "%s", set.Flags[i])
			}
			fmt.Fprintf(writer, " ;")
		}

		if set.Timeout != nil {
			fmt.Fprintf(writer, " timeout %d ;", int64(set.Timeout.Seconds()))
		}
		if set.GCInterval != nil {
			fmt.Fprintf(writer, " gc-interval %d ;", int64(set.GCInterval.Seconds()))
		}
		if set.Size != nil {
			fmt.Fprintf(writer, " size %d ;", *set.Size)
		}
		if set.Policy != nil {
			fmt.Fprintf(writer, " policy %s ;", *set.Policy)
		}
		if set.AutoMerge != nil && *set.AutoMerge {
			fmt.Fprintf(writer, " auto-merge ;")
		}

		if set.Comment != nil {
			fmt.Fprintf(writer, " comment %q ;", *set.Comment)
		}

		fmt.Fprintf(writer, " }")
	}

	fmt.Fprintf(writer, "\n")
}

// Object implementation for Map
func (mapObj *Map) validate(verb verb) error {
	if mapObj.Name == "" {
		return fmt.Errorf("no name specified for map")
	}

	switch verb {
	case addVerb:
		if (mapObj.Type == "" && mapObj.TypeOf == "") || (mapObj.Type != "" && mapObj.TypeOf != "") {
			return fmt.Errorf("map must specify either Type or TypeOf")
		}
		fallthrough
	case flushVerb:
		if mapObj.Handle != nil {
			return fmt.Errorf("cannot specify Handle in %s operation", verb)
		}
	case deleteVerb:
		// Handle can be nil or non-nil
	default:
		return fmt.Errorf("%s is not implemented for maps", verb)
	}

	return nil
}

func (mapObj *Map) writeOperation(verb verb, family Family, table string, writer io.Writer) {
	// Special case for delete-by-handle
	if verb == deleteVerb && mapObj.Handle != nil {
		fmt.Fprintf(writer, "delete map %s %s handle %d", family, table, *mapObj.Handle)
		return
	}

	fmt.Fprintf(writer, "%s map %s %s %s", verb, family, table, mapObj.Name)
	if verb == addVerb {
		fmt.Fprintf(writer, " {")

		if mapObj.Type != "" {
			fmt.Fprintf(writer, " type %s ;", mapObj.Type)
		} else {
			fmt.Fprintf(writer, " typeof %s ;", mapObj.TypeOf)
		}

		if len(mapObj.Flags) != 0 {
			fmt.Fprintf(writer, " flags ")
			for i := range mapObj.Flags {
				if i > 0 {
					fmt.Fprintf(writer, ",")
				}
				fmt.Fprintf(writer, "%s", mapObj.Flags[i])
			}
			fmt.Fprintf(writer, " ;")
		}

		if mapObj.Timeout != nil {
			fmt.Fprintf(writer, " timeout %d ;", int64(mapObj.Timeout.Seconds()))
		}
		if mapObj.GCInterval != nil {
			fmt.Fprintf(writer, " gc-interval %d ;", int64(mapObj.GCInterval.Seconds()))
		}
		if mapObj.Size != nil {
			fmt.Fprintf(writer, " size %d ;", *mapObj.Size)
		}
		if mapObj.Policy != nil {
			fmt.Fprintf(writer, " policy %s ;", *mapObj.Policy)
		}

		if mapObj.Comment != nil {
			fmt.Fprintf(writer, " comment %q ;", *mapObj.Comment)
		}

		fmt.Fprintf(writer, " }")
	}

	fmt.Fprintf(writer, "\n")
}

// Object implementation for Element
func (element *Element) validate(verb verb) error {
	if element.Name == "" {
		return fmt.Errorf("no set/map name specified for element")
	}

	return nil
}

func (element *Element) writeOperation(verb verb, family Family, table string, writer io.Writer) {
	fmt.Fprintf(writer, "%s element %s %s %s { %s", verb, family, table, element.Name, element.Key)

	if element.Value != "" {
		fmt.Fprintf(writer, " : %s", element.Value)
	}

	if verb == addVerb && element.Comment != nil {
		fmt.Fprintf(writer, " comment %q", *element.Comment)
	}

	fmt.Fprintf(writer, " }\n")
}