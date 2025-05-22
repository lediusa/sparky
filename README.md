## Sparky
     
###### a Reconnaissance tool for bug hunters to perform subdomain discovery, fuzzing, and vulnerability scanning.
     
#### Installation
```bash
go install github.com/lediusa/sparky/cmd/sparky@latest
sparky -id
```
     
#### Usage
```bash
sparky -d example.com -sm -vhost -jsf -wcd
sparky -f domains.txt -threads 5
```
     
#### Features

- Subdomain enumeration
- Virtual host discovery
- Smart fuzzing for 403/404 subdomains
- JavaScript fuzzing
- Web Cache Deception testing
- SQLi and Nuclei scanning
     
### License
MIT License
