variable "VERSION_ARG" {
  default = "0.0.0"
}

variable "REPO" {
  default = "artifact.aerospike.io/ecosystem-container-dev-local"
}

group "default" {
  targets = ["asvec"]
}

target "asvec" {
  context    = "."                             
  dockerfile = "./Dockerfile"
  platforms  = ["linux/amd64", "linux/arm64"]
  tags       = [
    "${REPO}/asvec:${VERSION_ARG}",
    "${REPO}/asvec:latest",
  ]
  args = {
    VERSION = "${VERSION_ARG}"
  }
}
