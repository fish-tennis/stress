
let reqhd = new Vue({
    el: "#createReq",
    data: {
        json: null,
        title:"",
        classList: [],
        pbmsgs: null,
        msgname: "",
        classname: "",
        reqname: "",
        show: false,
        showheader: true,
        showobj: null,
        options: {
            onChange: () => {
                //console.log("on change", this.json.hello)
            },
            onValidate: function (json) {

            },
            mode: 'text'
        },
    },
    methods: {
        onError(err) {
            // console.log("on err", err)
        },

        ensureCreateReq() {
            // console.log(this.classname, this.msgname, this.reqname)
            if (this.showobj === null) {
                config.createReq(this.msgname, this.classname, this.reqname, this.json)
            } else {
                let obj = this.showobj
                config.createReq(obj.item.req, obj.cls, obj.item.id, this.json)
            }

            this.show = false
        },

        onSelectMsg() {
            // console.log("onSelectMsg", this.msgname, this.json)
            config.client.GetInfo(this.msgname).then((rsp) => {
                this.json = rsp
            })
        },
        onSelectMsgClass() {
            if (this.classList.findIndex(ele => ele == this.classname) == -1) {
                this.classList.push(this.classname)
                // console.log("add")
            }

            // console.log("onSelectMsgClass", this.classname)
        },

        SetMsgs(msgs) {
            this.pbmsgs = msgs
            this.msgname = msgs[0]
            config.client.GetInfo(this.msgname).then((rsp) => {
                this.json = rsp
            })
        },
        SetReqs(reqs) {
            this.classList = Object.keys(reqs)
            this.classname = this.classList[0]
        },
        ShowByName(cls, item) {
            this.title = item.req
            this.showheader = false
            this.show = true
            this.showobj = { cls: cls, item: item }
            this.json = item.params
        },
        Show(cls) {
            this.classname = cls
            this.title = "创建请求"
            this.showheader = true
            this.show = true
            this.showobj = null
        }
    },

    mounted() {
        //    this.reqs = config.File["req"]
        //    console.log("on mounted", config.File)
    }
})