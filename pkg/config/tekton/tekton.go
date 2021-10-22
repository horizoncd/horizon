package tekton

type Mapper map[string]*Tekton

type Tekton struct {
	Server    string `yaml:"server"`
	Namespace string `yaml:"namespace"`
	S3        *S3    `yaml:"s3"`
}

type S3 struct {
	AccessKey        string `yaml:"accessKey"`
	SecretKey        string `yaml:"secretKey"`
	Region           string `yaml:"region"`
	Endpoint         string `yaml:"endpoint"`
	Bucket           string `yaml:"bucket"`
	SkipVerify       bool   `yaml:"skipVerify"`
	S3ForcePathStyle bool   `yaml:"s3ForcePathStyle"`
}
