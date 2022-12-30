let tabs = new Vue({
    el: "#tabs",
    data: {
        tabs: [],
        tabname: '',
        seq: 0,
    },
    methods: {
        current() {
            let idx = Number(this.tabname) -1
            return this.tabs[idx]
        },

        getById(id) {
            let tab = null
            for (let i=0; i< this.tabs.length; i++){
                if (this.tabs[i].id === id){
                    tab = this.tabs[i]
                    break
                }
            }
            return tab
        },

        buildTabItem (idx) {
            let name =  idx.toString()
            this.seq++
            return {
                id: this.seq,
                title: 'Tab ' + this.seq.toString(),
                name: name,
                logs:[],
                server:{
                    name: config.servername,
                    account: config.account,
                },
                statusBar:{
                    level: "status-name error",
                    status: "未登录",
                    info:"",
                },
            }
        },

        add(){
            let idx = this.tabs.length +1                
            let tab = this.buildTabItem(idx)
            this.tabs.push(tab)
            this.tabname = tab.name
            this.changeTab()
        },
        remove(name){
            if (this.tabs.length === 1){
                return
            }
            let idx = Number(name)-1
            // console.log(this.tabs[idx], idx)
            this.tabs.splice(idx, 1)
            if (idx >= this.tabs.length){
                idx--
            }

            this.tabs.forEach((item, idx)=>{
                idx++
                item.name = idx.toString()
                // item.title = 'Tab ' + item.name
            })
            this.tabname = this.tabs[idx].name
            this.changeTab()
        },

        onEdit(name, action) {
            if (action === 'remove') {
                this.remove(name)
                return 
            }

            if (action === 'add'){
                this.add()
            }
        },
        onClickTab(ins) {
            // console.log(this.current())
            this.changeTab()
        },
        changeTab(){
            let tab = this.current()
            logbox.setData(tab.logs)
            bar.setData(tab.statusBar)
            config.servername = tab.server.name
            config.account = tab.server.account
        },
        setServer(){
            let tab = this.current()
            tab.server.name = config.servername
            tab.server.account = config.account
            // console.log(this.tabs)
        }
    },
    mounted() {
        this.add()
    }
})