package main

var platforms = []string{
	"ec2",
	"vsphere",
}

func main() {
	// for _, platform := range platforms {
	// 	tfVarsFile := filepath.Join(platform, "templates", "terraform.tfvars")

	// 	tfVars, err := provisioner.GetTFVariables(platform, map[string]interface{}{})
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	if err := ioutil.WriteFile(tfVarsFile, tfVars, 0644); err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("saved terraform variables for %s to %s\n", platform, tfVarsFile)
	// }
}
