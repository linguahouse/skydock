// this plugin is seeting names based on the rancher labels

function objdump(arr, level) {
    var dumped_text = "";
    if(!level) level = 0;

    var level_padding = "";
    for(var j=0;j<level+1;j++) level_padding += "    ";

    if(typeof(arr) == 'object') {
        for(var item in arr) {
            var value = arr[item];

            if(typeof(value) == 'object') {
                dumped_text += level_padding + "'" + item + "' ...\n";
                dumped_text += objdump(value,level+1);
            } else {
                dumped_text += level_padding + "'" + item + "' => \"" + value + "\"\n";
            }
        }
    } else {
        dumped_text = "===>"+arr+"<===("+typeof(arr)+")";
    }
    return dumped_text;
}


function createService(container) {
    var port = getDefaultPort(container);
    var address = container.NetworkSettings.IpAddress;
    
    console.log("Rancher plugin in use");

    if (container.Config.Labels['io.rancher.container.ip']){
        var rancherAddress=container.Config.Labels['io.rancher.container.ip'];
        console.log("Detected rancher address: "+rancherAddress);
        var slashpos = rancherAddress.indexOf("/");
        if (slashpos){
            address=rancherAddress.substring(0,slashpos);
        }else{
            address=rancherAddress;
        }
        console.log("Address to be used is "+address);
    }
    
    return {
        Port: port,
        Environment: defaultEnvironment,
        TTL: defaultTTL,
        Service: cleanImageName(container.Image),
        Instance: removeSlash(container.Name),
        Host: address
    }; 
}

function getDefaultPort(container) {
    // if we have any exposed ports use those
    var port = 0;
    var ports = container.NetworkSettings.Ports;
    if (Object.keys(ports).length > 0) {
        for (var key in ports) {
            var value = ports[key];
            if (value !== null && value.length > 0) {
                for (var i = 0; i < value.length; i++) {
                    var hp = parseInt(value[i].HostPort);
                    if (port === 0 || hp < port) {
                        port = hp; 
                    }
                }  
            } else if (port === 0) {
                // just grab the key value 
                var expose = parseInt(key.split("/")[0]); 
                port = expose;
            }
        }
    } 
     
    if (port === 0) {
        port = 80; 
    }
    return port;
}
