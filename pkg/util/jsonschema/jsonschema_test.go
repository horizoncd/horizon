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
		"resource": {
          "type": "string",
          "title": "规格",
          "description": "应用上建议选择tiny或者small规格（测试环境集群自动继承，节省资源使用），线上集群可选大规格"
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

func TestDependency(t *testing.T) {
	schema := `{
    "properties": {
        "app": {
            "properties": {
                "health": {
                    "dependencies": {
                        "check": [
                            "port"
                        ],
                        "offline": [
                            "port"
                        ],
                        "online": [
                            "port"
                        ],
                        "status": [
                            "port"
                        ]
                    },
                    "properties": {
                        "check": {
                            "description": "存活状态会在应用运行期间检测应用健康情况，检测失败时会对应用进行重启。接口如: /api/test",
                            "pattern": "^/.*$",
                            "title": "存活状态",
                            "type": "string"
                        },
                        "expectedStartTime": {
                            "default": 200,
                            "description": "该时间必须为10的整数倍。会根据该期望启动时间调整健康检查的重试次数，在该时间内健康检查不通过，应用会启动失败",
                            "minimum": 30,
                            "multipleOf": 10,
                            "title": "期望启动时间（单位秒）",
                            "type": "number"
                        },
                        "offline": {
                            "description": "下线接口会在应用停止之前进行调用，如果调用失败，则忽略。接口如: /health/offline",
                            "pattern": "^/.*$",
                            "title": "下线",
                            "type": "string"
                        },
                        "online": {
                            "description": "上线接口会在应用启动之后进行调用，如果调用失败，则应用启动失败。接口如: /health/online",
                            "pattern": "^/.*$",
                            "title": "上线",
                            "type": "string"
                        },
                        "port": {
                            "maximum": 65535,
                            "minimum": 1024,
                            "type": "integer"
                        },
                        "status": {
                            "description": "就绪状态会在应用运行期间检测应用是否处于上线状态，检测失败时显示下线状态。接口如: /health/status",
                            "pattern": "^/.*$",
                            "title": "就绪状态",
                            "type": "string"
                        }
                    },
                    "title": "健康检查",
                    "type": "object"
                },
                "params": {
                    "properties": {
                        "jvmExtra": {
                            "description": "如果集群标签中也有jvmExtra，此处的jvmExtra会和标签中的jvmExtra合并生效，如有重复，集群标签中的jvmExtra优先级更高。` +
		`（通过Overmind创建的集群会在标签中传入jvmExtra）",
                            "type": "string"
                        },
                        "mainClassName": {
                            "type": "string"
                        },
                        "xdebugAddress": {
                            "pattern": "^\\d*$",
                            "type": "string"
                        },
                        "xms": {
                            "default": "512",
                            "pattern": "^\\d*$",
                            "type": "string"
                        },
                        "xmx": {
                            "default": "1024",
                            "pattern": "^\\d*$",
                            "type": "string"
                        }
                    },
                    "required": [
                        "mainClassName"
                    ],
                    "title": "参数",
                    "type": "object"
                },
                "spec": {
                    "dependencies": {
                        "resource": {
                            "oneOf": [
                                {
                                    "properties": {
                                        "cpu": {
                                            "default": 500,
                                            "description": "单位：m，应用上建议选择500或者1000规格（测试环境集群自动继承，节省资源使用），线上集群可选大规格",
                                            "enum": [
                                                500,
                                                1000,
                                                2000,
                                                4000,
                                                6000,
                                                8000
                                            ],
                                            "title": "cpu",
                                            "type": "integer"
                                        },
                                        "memory": {
                                            "default": 1024,
                                            "description": "单位：MB，应用上建议选择1024或者2048规格（测试环境集群自动继承，节省资源使用），线上集群可选大规格",
                                            "enum": [
                                                1024,
                                                2048,
                                                4096,
                                                6144,
                                                8192,
                                                10240,
                                                12288,
                                                14336,
                                                16384
                                            ],
                                            "title": "memory",
                                            "type": "integer"
                                        },
                                        "resource": {
                                            "enum": [
                                                "flexible"
                                            ],
                                            "title": "flexible"
                                        }
                                    }
                                },
                                {
                                    "properties": {
                                        "resource": {
                                            "enum": [
                                                "tiny"
                                            ]
                                        }
                                    }
                                },
                                {
                                    "properties": {
                                        "resource": {
                                            "enum": [
                                                "x-small"
                                            ]
                                        }
                                    }
                                },
                                {
                                    "properties": {
                                        "resource": {
                                            "enum": [
                                                "small"
                                            ]
                                        }
                                    }
                                },
                                {
                                    "properties": {
                                        "resource": {
                                            "enum": [
                                                "middle"
                                            ]
                                        }
                                    }
                                },
                                {
                                    "properties": {
                                        "resource": {
                                            "enum": [
                                                "large"
                                            ]
                                        }
                                    }
                                }
                            ]
                        }
                    },
                    "properties": {
                        "replicas": {
                            "default": 1,
                            "maximum": 30,
                            "minimum": 0,
                            "title": "副本数",
                            "type": "integer"
                        },
                        "resource": {
                            "default": "x-small",
                            "description": "应用上建议选择tiny或者small规格（测试环境集群自动继承，节省资源使用），线上集群可选大规格",
                            "oneOf": [
                                {
                                    "enum": [
                                        "tiny"
                                    ],
                                    "title": "tiny(0.5C1G)"
                                },
                                {
                                    "enum": [
                                        "x-small"
                                    ],
                                    "title": "x-small(1C2G)"
                                },
                                {
                                    "enum": [
                                        "small"
                                    ],
                                    "title": "small(2C4G)"
                                },
                                {
                                    "enum": [
                                        "middle"
                                    ],
                                    "title": "middle(4C8G)"
                                },
                                {
                                    "enum": [
                                        "flexible"
                                    ],
                                    "title": "flexible"
                                }
                            ],
                            "title": "规格",
                            "type": "string"
                        }
                    },
                    "title": "规格",
                    "type": "object"
                },
                "strategy": {
                    "properties": {
                        "pauseType": {
                            "default": "all",
                            "oneOf": [
                                {
                                    "enum": [
                                        "first"
                                    ],
                                    "title": "第一批暂停"
                                },
                                {
                                    "enum": [
                                        "all"
                                    ],
                                    "title": "全部暂停"
                                },
                                {
                                    "enum": [
                                        "none"
                                    ],
                                    "title": "全不暂停"
                                }
                            ],
                            "title": "暂停策略",
                            "type": "string"
                        },
                        "stepsTotal": {
                            "default": 1,
                            "enum": [
                                1,
                                2,
                                3,
                                4,
                                5
                            ],
                            "title": "发布批次（多批次情况下，第一批默认为1个实例）",
                            "type": "integer"
                        }
                    },
                    "title": "发布策略",
                    "type": "object"
                }
            },
            "title": "应用",
            "type": "object"
        },
        "memcached": {
            "dependencies": {
                "enabled": {
                    "oneOf": [
                        {
                            "properties": {
                                "enabled": {
                                    "enum": [
                                        true
                                    ]
                                },
                                "size": {
                                    "default": "64M",
                                    "description": "选择合适的memcache规格，memcache可用内存为规格的3/4, 测试环境选择最小规格",
                                    "oneOf": [
                                        {
                                            "enum": [
                                                "64M"
                                            ],
                                            "title": "64M(memory 64M,cpu 50m)"
                                        },
                                        {
                                            "enum": [
                                                "1G"
                                            ],
                                            "title": "1G(memory 1024M, cpu 1000m)"
                                        },
                                        {
                                            "enum": [
                                                "2G"
                                            ],
                                            "title": "2G(memory 2048M, cpu 1000m)"
                                        },
                                        {
                                            "enum": [
                                                "4G"
                                            ],
                                            "title": "4G(memory 4096M, cpu 1000m)"
                                        },
                                        {
                                            "enum": [
                                                "8G"
                                            ],
                                            "title": "8G(memory 8192M, cpu 1000m)"
                                        },
                                        {
                                            "enum": [
                                                "16G"
                                            ],
                                            "title": "16G(memory 16384M, cpu 2000m)"
                                        }
                                    ],
                                    "title": "选择规格",
                                    "type": "string"
                                }
                            }
                        },
                        {
                            "properties": {
                                "enabled": {
                                    "enum": [
                                        false
                                    ]
                                }
                            }
                        }
                    ]
                }
            },
            "properties": {
                "enabled": {
                    "default": false,
                    "description": "是否使用本地memcached缓存组件(曲库相关应用), memcached访问端口11215, memcached metric端口9150",
                    "enum": [
                        true,
                        false
                    ],
                    "title": "enabled",
                    "type": "boolean"
                }
            },
            "required": [
                "enabled"
            ],
            "title": "本地memcached",
            "type": "object"
        }
    },
    "type": "object"
}`
	document := `{
    "app": {
        "health": {
            "expectedStartTime": 210
        },
        "params": {
            "jvmExtra": "-D server.port=8888",
            "mainClassName": "org.springframework.boot.loader.JarLauncher",
            "xms": "512",
            "xmx": "1024"
        },
        "spec": {
            "cpu": 2000,
            "memory": 4096,
            "replicas": 3,
            "resource": "flexible"
        },
        "strategy": {
            "pauseType": "all",
            "stepsTotal": 3
        }
    },
    "memcached": {
        "enabled": false
    }
}`

	// 1. cpu allowed when resource is flexible
	err := Validate(schema, document)
	assert.Nil(t, err)

	// 2. cpu not allowed when resource is not flexible
	document = `{
    "app": {
        "health": {
            "expectedStartTime": 210
        },
        "params": {
            "jvmExtra": "-D server.port=8888",
            "mainClassName": "org.springframework.boot.loader.JarLauncher",
            "xms": "512",
            "xmx": "1024"
        },
        "spec": {
            "cpu": 2000,
            "memory": 4096,
            "replicas": 3,
            "resource": "x-small"
        },
        "strategy": {
            "pauseType": "all",
            "stepsTotal": 3
        }
    },
    "memcached": {
        "enabled": false
    }
}`
	err = Validate(schema, document)
	assert.NotNil(t, err)

	// 3. additional field not allowed
	document = `{
    "app": {
		"kkk": 123,
        "health": {
            "expectedStartTime": 210
        },
        "params": {
            "jvmExtra": "-D server.port=8888",
            "mainClassName": "org.springframework.boot.loader.JarLauncher",
            "xms": "512",
            "xmx": "1024"
        },
        "spec": {
            "cpu": 2000,
            "memory": 4096,
            "replicas": 3,
            "resource": "flexible"
        },
        "strategy": {
            "pauseType": "all",
            "stepsTotal": 3
        }
    },
    "memcached": {
        "enabled": false
    }
}`
	err = Validate(schema, document)
	assert.NotNil(t, err)
}
