package jsonschema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	schema := `{
  "type": "object",
  "properties": {
    "app": {
      "title": "App Config",
      "type": "object",
      "properties": {
        "params": {
          "title": "参数",
          "type": "object",
          "properties": {
            "mainClassName": {
              "type": "string"
            },
            "xmx": {
              "type": "string",
              "default": "512"
            },
            "xms": {
              "type": "string",
              "default": "512"
            },
            "maxPerm": {
              "type": "string",
              "default": "128"
            },
            "xdebugAddress": {
              "type": "string"
            },
            "jvmExtra": {
              "type": "string"
            }
          },
          "required": ["mainClassName"]
        },
        "health": {
          "title": "健康检查",
          "type": "object",
          "properties": {
            "port": {
              "type": "integer"
            },
            "lifecycle":{
              "title": "优雅启停",
              "type": "object",
              "properties": {
                "online":{
                  "title": "上线",
                  "description": "上线接口会在应用启动之后进行调用，如果调用失败，则应用启动失败",
                  "$ref": "#/$defs/lifecycle"
                },
                "offline":{
                  "title": "下线",
                  "description": "下线接口会在应用停止之前进行调用，如果调用失败，则忽略",
                  "$ref": "#/$defs/lifecycle"
                }
              }
            },
            "probe":{
              "title": "健康检查",
              "type": "object",
              "properties": {
                "check":{
                  "title": "存活状态",
                  "description": "存活状态会在应用运行期间检测应用健康情况，检测失败时会对应用进行重启",
                  "$ref": "#/$defs/probe"
                },
                "status":{
                  "title": "就绪状态",
                  "description": "就绪状态会在应用运行期间检测应用是否处于上线状态，检测失败时显示下线状态",
                  "$ref": "#/$defs/probe"
                }
              }
            }
          },
          "dependencies": {
            "lifecycle": ["port"],
            "probe": ["port"]
          }
        }
      }
    }
  },

  "$defs": {
    "lifecycle": {
      "type": "object",
      "properties": {
        "url": {
          "title": "接口",
          "description": "接口路径",
          "type": "string"
        },
        "timeoutSeconds": {
          "title": "超时时间",
          "description": "请求接口的超时时间(单位为s)",
          "type": "integer"
        },
        "periodSeconds": {
          "title": "检测周期",
          "description": "连续两次检测之间的时间间隔(单位为s)",
          "type": "integer"
        },
        "retry": {
          "title": "重试次数",
          "description": "检测失败后重试的次数",
          "type": "integer"
        }
      }
    },
    "probe": {
      "type": "object",
      "properties": {
        "url": {
          "title": "接口",
          "description": "接口路径",
          "type": "string"
        },
        "initialDelaySeconds": {
          "title": "延迟启动",
          "description": "应用启动延迟等待该时间再进行检测",
          "type": "integer"
        },
        "timeoutSeconds": {
          "title": "超时时间",
          "description": "请求接口的超时时间(单位为s)",
          "type": "integer"
        },
        "periodSeconds": {
          "title": "重试次数",
          "description": "连续两次检测之间的时间间隔(单位为s)",
          "type": "integer"
        },
        "failureThreshold": {
          "title": "失败阈值",
          "description": "连续检测失败超过该次数，才认为最终检测失败",
          "type": "integer"
        }
      }
    }
  }
}
`
	document := `{
        "app":{
            "params":{
                "xmx":"512",
                "xms":"512",
                "maxPerm":"128",
                "mainClassName":"com.netease.horizon.WebApplication",
                "jvmExtra":"-Dserver.port=8080"
            },
            "resource":"x-small",
            "health":{
                "lifecycle":{
                    "online":{
                        "url":"/online",
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "retry":20
                    },
                    "offline":{
                        "url":"/offline",
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "retry":20
                    }
                },
                "probe":{
                    "check":{
                        "url":"/check",
                        "initialDelaySeconds":200,
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "failureThreshold":3
                    },
                    "status":{
                        "url":"/status",
                        "initialDelaySeconds":200,
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "failureThreshold":3
                    }
                },
                "port":8080
            }
        }
    }`

	// 1. normal
	err := Validate(schema, document)
	assert.Nil(t, err)

	// 2. error
	document = `{
        "app":{
            "params":{
                "xmx":"512",
                "xms":"512",
                "maxPerm":"128",
                "jvmExtra":"-Dserver.port=8080"
            },
            "resource":"x-small",
            "health":{
                "lifecycle":{
                    "online":{
                        "url":"/online",
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "retry":20
                    },
                    "offline":{
                        "url":"/offline",
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "retry":20
                    }
                },
                "probe":{
                    "check":{
                        "url":"/check",
                        "initialDelaySeconds":200,
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "failureThreshold":3
                    },
                    "status":{
                        "url":"/status",
                        "initialDelaySeconds":200,
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "failureThreshold":3
                    }
                },
                "port":8080
            }
        }
    }`

	err = Validate(schema, document)
	assert.NotNil(t, err)
	t.Logf("%v", err)

	// 3. error
	document = `{
        "app":{
            "params":{
                "xmx":"512",
                "xms":"512",
                "maxPerm":"128",
			    "mainClassName":"com.netease.horizon.WebApplication",
                "jvmExtra":"-Dserver.port=8080"
            },
            "resource":"x-small",
            "health":{
                "lifecycle":{
                    "online":{
                        "url":"/online",
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "retry":20
                    },
                    "offline":{
                        "url":"/offline",
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "retry":20
                    }
                },
                "probe":{
                    "check":{
                        "url":"/check",
                        "initialDelaySeconds":200,
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "failureThreshold":3
                    },
                    "status":{
                        "url":"/status",
                        "initialDelaySeconds":200,
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "failureThreshold":3
                    }
                }
            }
        }
    }`

	err = Validate(schema, document)
	assert.NotNil(t, err)
	t.Logf("%v", err)

	// 4. map[string]interface{} type
	document = `{
        "app":{
            "params":{
                "xmx":"512",
                "xms":"512",
                "maxPerm":"128",
			    "mainClassName":"com.netease.horizon.WebApplication",
                "jvmExtra":"-Dserver.port=8080"
            },
            "resource":"x-small",
            "health":{
                "lifecycle":{
                    "online":{
                        "url":"/online",
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "retry":20
                    },
                    "offline":{
                        "url":"/offline",
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "retry":20
                    }
                },
                "probe":{
                    "check":{
                        "url":"/check",
                        "initialDelaySeconds":200,
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "failureThreshold":3
                    },
                    "status":{
                        "url":"/status",
                        "initialDelaySeconds":200,
                        "timeoutSeconds":3,
                        "periodSeconds":15,
                        "failureThreshold":3
                    }
                },
				"port":8080
            }
        }
    }`
	var documentMap map[string]interface{}
	err = json.Unmarshal([]byte(document), &documentMap)
	assert.Nil(t, err)
	err = Validate(schema, documentMap)
	assert.Nil(t, err)

	// unsupported type
	var documentMapMap map[string]map[string]interface{}
	err = json.Unmarshal([]byte(document), &documentMapMap)
	assert.Nil(t, err)
	err = Validate(schema, documentMapMap)
	assert.NotNil(t, err)
}
