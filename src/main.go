package main
import(
    "syscall"
    //"bufio"
    //"os"
    "net"
    "fmt"
    "time"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
    "flag"
)

var traceip string
var maxttl int
func init(){
	flag.StringVar(&traceip, "ip", "", "trace dest ipaddress(ipv4)", )
	flag.IntVar(&maxttl, "maxttl", 32, "max ttl value default 32")
}
func main(){
	flag.Parse()
	if(traceip == ""){
		fmt.Print("must have ip flag!")
		return
	}
	var channel = make(chan struct { seq int; code int; sourecip string})
	var processid uint16 = 16;
    fmt.Printf("start test! destip:%s\n", traceip)
    var dstipinfo, error2 =  net.ResolveIPAddr("ip4", traceip);
    if(error2 != nil){
    	fmt.Print("parse ip error")
    	return
	}
	var address, error = net.ResolveIPAddr("ip4", "104.192.82.125");
	if(error != nil){
			fmt.Print(error)
			return
	}
	con, error := net.ListenIP("ip4:icmp", address)
	if(error != nil){
			fmt.Print(error)
			return
	}
	con.SetDeadline(time.Now().Add(10 * time.Second))

    go func(){
		var recvBuff = make([]byte, 1024)
		for{
			packLength, addr, error := con.ReadFrom(recvBuff)
			if(packLength == 0 || error != nil){
				fmt.Print(error)
				break
			}
			var realpack = recvBuff[:packLength]
			var icmppack = gopacket.NewPacket(realpack, layers.LayerTypeICMPv4, gopacket.Default)
			var layer = icmppack.Layer(layers.LayerTypeICMPv4)
			var icmpdata, _ = layer.(*layers.ICMPv4)

			//fmt.Printf("id: %d   seqid:%d   type:%d    code:%d\n", icmpdata.Id, icmpdata.Seq, icmpdata.TypeCode.Type(), icmpdata.TypeCode.Code())
			if(icmpdata.TypeCode.Type() == 11){
				//fmt.Printf("playload:%X\n", icmpdata.Payload)
				//var ipPack, _ = gopacket.NewPacket(icmpdata.Payload, layers.LayerTypeIPv4, gopacket.Default).Layer(layers.LayerTypeIPv4).(*layers.IPv4)
				var icmptimeout, _ = gopacket.NewPacket(icmpdata.Payload[20:], layers.LayerTypeICMPv4, gopacket.Default).Layer(layers.LayerTypeICMPv4).(*layers.ICMPv4)
				//fmt.Printf("timeoutpackid:%d sourceip:%s\n", icmptimeout.Id, addr.String())
				channel <- struct {
					seq  int
					code int
					sourecip string
				}{ int(icmptimeout.Seq),  int(icmpdata.TypeCode.Type()), addr.String()}
			}
			if(icmpdata.TypeCode.Type() == 0){
				channel <- struct {
					seq  int
					code int
					sourecip string
				}{int(icmpdata.Seq), int(icmpdata.TypeCode.Type()), addr.String()}
				break
			}
			//fmt.Printf("%s:%s, %X\n", addr.Network(), addr.String(), recvBuff[:packLength])
		}
    }()

    Loop:
	for i := 1; i <= maxttl; i++ {
		sendicmp(processid, address ,dstipinfo, uint8(i))
		select{
			case resu :=  <-channel:
				fmt.Printf("recv result code:%d, seq:%d  sourceip:%s\n", resu.code, resu.seq, resu.sourecip)
				if(resu.code == 0){
					fmt.Print("trace success!\n")
					break Loop
				}
			case <- time.After(1 * time.Second):
				fmt.Print("recv timeout\n")
		}
	}

    //reader := bufio.NewReader(os.Stdin)
    //reader.ReadByte()
}

func sendicmp(processid uint16, address *net.IPAddr, dstipinfo *net.IPAddr, ttl uint8){
	var icmppacket = &layers.ICMPv4{
		Seq:uint16(ttl),
		Id:processid,
		TypeCode:layers.CreateICMPv4TypeCode(8, 0),
	}
	var ippacket = &layers.IPv4{
		SrcIP:    address.IP,
		DstIP:    dstipinfo.IP,
		Protocol: layers.IPProtocolICMPv4,
		TTL:ttl,
		Version:4,
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	gopacket.SerializeLayers(buf, opts, ippacket, icmppacket, gopacket.Payload([]byte{1, 2, 3, 4}))
	var t = buf.Bytes()
	//fmt.Printf("_%X_", t)
	//_, error = con.WriteTo(t,  &net.IPAddr{ net.IPv4(14,215,177,39), "" })
	fd, error := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if(error != nil){
		fmt.Print("create raw socket error!\n")
		fmt.Print(error)
	}
	var fixaddr [4]byte;
	copy(fixaddr[:], dstipinfo.IP[12:16])
	error = syscall.Sendto(fd, t, 0, &syscall.SockaddrInet4{Addr:fixaddr, Port:0})
	if error != nil{
		fmt.Print("sento error")
		fmt.Print(error)
	}
}
