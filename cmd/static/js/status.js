let bar = new Vue({
    el: "#statusBar",
    data: {
        bar:{
            level: "status-name error",
            status: "未登录",
            info:"",
        },
    },
    methods: {
        changeStatus(data){
            let bar = tabs.current().statusBar
            if (data.tab != 0) {
                bar = tabs.getById(data.tab).statusBar
            }
            // console.log('changeStatus', data, this.bar)
            bar.level = `status-name ${data.status}`
            bar.status = data.name
            bar.info = data.info
            
        },
        changeLevel(lv){
            let bar = tabs.current().statusBar
            bar.level = `status-name ${lv}`
            bar.info = bar.info.replace('在线','离线')
        },
        setData(data){
            this.bar = data
        }
    },

    mounted(){
        GOAgent.Reg('info.status', this.changeStatus)
    }
})