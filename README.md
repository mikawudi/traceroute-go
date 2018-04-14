### sample:  
main -ip 8.8.8.8

### result:  
/**  same info **/  
recv result code:11, seq:7  sourceip:213.248.100.57  
recv result code:11, seq:8  sourceip:108.170.242.241  
recv result code:11, seq:9  sourceip:209.85.240.43  
recv result code:0, seq:10  sourceip:8.8.8.8  
trace success!  

### note:  
&ensp;&ensp;code is icmptype field, seq is icmp seq field (it's same as ttl, you can edit code to change it), sourceip is recv icmp ack from where, also ,you can use -maxttl to change maxttl value  
icmp id field is 16 ....it's just a magic number

#### todo
&ensp;&ensp;add delay time, send multi icmp pack for every ttl to get package 
loss rate, and so on....