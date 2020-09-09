package agent

import (
	"fmt"
	"strings"

	"arhat.dev/aranya-proto/aranyagopb"
	"arhat.dev/arhat/pkg/conf"
)

func convertNodeExtInfo(extInfo []conf.ArhatNodeExtInfo) ([]*aranyagopb.NodeExtInfo, error) {
	extInfoValueTypeMap := map[string]aranyagopb.NodeExtInfo_ExtInfoValueType{
		"":       aranyagopb.NODE_EXT_INFO_TYPE_STRING,
		"string": aranyagopb.NODE_EXT_INFO_TYPE_STRING,
		"int":    aranyagopb.NODE_EXT_INFO_TYPE_INTEGER,
		"float":  aranyagopb.NODE_EXT_INFO_TYPE_FLOAT,
	}
	extInfoOperatorMap := map[string]aranyagopb.NodeExtInfo_ExtInfoOperator{
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
		case aranyagopb.NODE_EXT_INFO_TYPE_INTEGER, aranyagopb.NODE_EXT_INFO_TYPE_FLOAT:
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

		result = append(result, &aranyagopb.NodeExtInfo{
			Value:     value,
			ValueType: valueType,
			Operator:  operator,
			ApplyTo:   info.ApplyTo,
		})
	}

	return result, nil
}
