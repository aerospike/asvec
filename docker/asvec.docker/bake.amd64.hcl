
group "default" {
  targets = ["asvec"]
}

target "asvec" {
  context    = "."                             
  dockerfile = "./Dockerfile"
  platforms  = ["linux/amd64"]
  tags       = [
  ]
}
