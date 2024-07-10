TAG=0.1.0
    
    set -e
    gh release list -R github.com/aerospike/asvec -L 100 |grep Pre-release |awk -F'\t' '{print $3}' |while read line
    do
    if [ "$line" != "${TAG}" ]
    then
        if [[ $line =~ ^${TAG}- ]]
        then
        echo "Removing $line"
        gh release delete $line -R github.com/aerospike/asvec --yes --cleanup-tag
        fi
    fi
    done
