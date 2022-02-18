package main

func main() {
	//httpClient := http.CreateHttpClient()

	// test get
	//body, err := httpClient.Get(tools.GetGuid(), "http://www.baidu.com")
	//if err != nil {
	//    fmt.Println(err)
	//    return
	//}
	//fmt.Println(body)

	// test post
	//request := ecs.QueryServerDetailRequest{
	//    ServerId:"62620838-c331-460e-bce1-648a7e49ce84",
	//}
	//request.Account = "2017徐小白来了"
	//request.DataCenter = "bj_02"
	//response, err := httpClient.Post(tools.GetGuid(), "http://vpc-biz.jcloud.com/compute?Action=queryServerDetail", request, nil)
	//if err != nil {
	//    fmt.Println(err)
	//    return
	//}
	//fmt.Println(response)
	//
	//// test post
	//request2 := ecs.QueryServerDetailRequest{
	//    ServerId:"62620838-c331-460e-bce1-648a7e49ce84",
	//}
	//request2.Account = "2017徐小白来了"
	//request2.DataCenter = "bj_02"
	//response2 := &ecs.QueryServerDetailResponse{}
	//err = httpClient.PostJson(tools.GetGuid(), "http://vpc-biz.jcloud.com/compute?Action=queryServerDetail", request2, response2, nil)
	//if err != nil {
	//    fmt.Println(err)
	//    return
	//}
	//fmt.Println(response2)
}
