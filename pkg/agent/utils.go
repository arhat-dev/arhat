/*
Copyright 2020 The arhat.dev Authors.

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

package agent

import (
	"fmt"
	"strings"

	"arhat.dev/aranya-proto/aranyagopb"

	"arhat.dev/arhat/pkg/conf"
)

func convertNodeExtInfo(extInfo []conf.NodeExtInfo) ([]*aranyagopb.NodeExtInfo, error) {
	extInfoValueTypeMap := map[string]aranyagopb.NodeExtInfo_ValueType{
		"":       aranyagopb.NODE_EXT_INFO_TYPE_STRING,
		"string": aranyagopb.NODE_EXT_INFO_TYPE_STRING,
		"number": aranyagopb.NODE_EXT_INFO_TYPE_NUMBER,
	}
	extInfoOperatorMap := map[string]aranyagopb.NodeExtInfo_Operator{
		"":   aranyagopb.NODE_EXT_INFO_OPERATOR_SET,
		"=":  aranyagopb.NODE_EXT_INFO_OPERATOR_SET,
		"+=": aranyagopb.NODE_EXT_INFO_OPERATOR_ADD,
		"-=": aranyagopb.NODE_EXT_INFO_OPERATOR_MINUS,
	}

	var result []*aranyagopb.NodeExtInfo
	for _, info := range extInfo {
		operator, ok := extInfoOperatorMap[strings.ToLower(info.Operator)]
		if !ok {
			return nil, fmt.Errorf("unsupported operator %q", info.Operator)
		}
		valueType, ok := extInfoValueTypeMap[strings.ToLower(info.ValueType)]
		if !ok {
			return nil, fmt.Errorf("unsupported valueType %q", info.ValueType)
		}

		switch valueType {
		case aranyagopb.NODE_EXT_INFO_TYPE_STRING:
			switch operator {
			case aranyagopb.NODE_EXT_INFO_OPERATOR_SET,
				aranyagopb.NODE_EXT_INFO_OPERATOR_ADD:
				// ok
			default:
				return nil, fmt.Errorf("valueType string do not support operator %q", info.Operator)
			}
		case aranyagopb.NODE_EXT_INFO_TYPE_NUMBER:
			switch operator {
			case aranyagopb.NODE_EXT_INFO_OPERATOR_SET,
				aranyagopb.NODE_EXT_INFO_OPERATOR_ADD,
				aranyagopb.NODE_EXT_INFO_OPERATOR_MINUS:
				// ok
			default:
				return nil, fmt.Errorf("valueType %q does not support operator %q", info.ValueType, info.Operator)
			}
		default:
			return nil, fmt.Errorf("unsupported vlaueType %q", info.ValueType)
		}

		value, err := info.ValueFrom.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to get value ext info value: %w", err)
		}

		var (
			target    aranyagopb.NodeExtInfo_Target
			targetKey string
		)
		switch {
		case strings.HasPrefix(info.ApplyTo, `metadata.annotations['`):
			target = aranyagopb.NODE_EXT_INFO_TARGET_ANNOTATION
			targetKey = strings.TrimPrefix(info.ApplyTo, `metadata.annotations['`)
		case strings.HasPrefix(info.ApplyTo, `metadata.labels['`):
			target = aranyagopb.NODE_EXT_INFO_TARGET_LABEL
			targetKey = strings.TrimPrefix(info.ApplyTo, `metadata.labels['`)
		default:
			return nil, fmt.Errorf("invalid ext info target")
		}
		targetKey = strings.TrimSuffix(targetKey, `']`)

		result = append(result, &aranyagopb.NodeExtInfo{
			Value:     value,
			ValueType: valueType,
			Operator:  operator,
			Target:    target,
			TargetKey: targetKey,
		})
	}

	return result, nil
}
