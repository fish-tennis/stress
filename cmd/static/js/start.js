let treetpl = `
<div class="button-group fill">
<button class="view {{item.id}}" @click=showBtree(item)>{{item.title}}</button>
<button class="button" @click=run(item)><i class="fa fa-play" aria-hidden="true"></i></button>
<button class="button" @click=stress(item)><i class="fa fa-rocket" aria-hidden="true"></i></button>
</div>
`

Vue.component('treebtn', {
    props: ['item'],
    methods: {
        run(tree) {
            stressDialog.RunTree(tree)
        },
        stress(tree) {
            stressDialog.Show(tree)
        },
        showBtree(tree){
            // console.log('show', tree)
            config.setSelectedTree(tree.id)
            btree.Show()
        }
    },
    template: treetpl
})

let stressDialog = new Vue({
    el: "#stress",
    data: {
        robot: new RobotAgent(),
        show: false,
        form: {
            start: 1,
            count: 1,
        },
        title: "压力测试",
    },
    methods: {
        onClick() {
            let sinfo = {
                tree_id: this.form.tree,
                start: Number(this.form.start),
                count: Number(this.form.count),
            }

            let server = config.GetServerConfig()
            // console.log('stress', sinfo, server)

            this.robot.StressBehaviorTree(JSON.stringify(server), JSON.stringify(sinfo)).
                then(() => {
                    config.saveLastloginServer()
                }).catch((e) => {
                    this.$notify({
                        title: '失败',
                        message: '运行失败:' + e,
                        position: 'top-left',
                        duration: 2000,
                        // offset: 100,
                        type: 'error'
                    });
                })
            this.show = false

        },

        RunTree(tree){
            let tab = tabs.current()
            stressDialog.robot.RunBehaviorTree(tree.id, tab.id, JSON.stringify(config.GetServerConfig()))
                .then(() => {
                    config.saveLastloginServer()
                }).catch((e) => {
                    this.$notify({
                        title: '失败',
                        message: '运行失败:' + e,
                        position: 'top-left',
                        duration: 2000,
                        // offset: 100,
                        type: 'error'
                    });
                })
        },

        Show(tree) {
            this.show = true
            this.form['server'] = config.servername
            this.form['tree'] = tree.id
            this.form['treeName'] = tree.title
        }
    }
})

let zoneDialog = new Vue({
    el: "#zone",
    data: {
        show:false,
        radio:'',
        zones:[],
        zone:null,
        title:"选择区服",
    },

    methods:{
        onClick(){
            // console.log(this.zone)
            config.saveZone(this.zone.id)
            this.show = false
        },
        onChange(item){
            // console.log("change", item)
            this.zone = item
        },
        Show(){
            GOAgent.Call("FetchZones", JSON.stringify(config.GetServerConfig()), "").then((rsp)=>{
                // console.log("zones", rsp)
                this.zones = rsp
                if (rsp.length !==0 ){
                    this.zone = rsp[0]
                    this.radio = this.zone.name
                }
                this.show = true
            }).catch((e)=>{
                this.$notify({
                    title: '拉取区服信息失败',
                    message: '运行失败:' + e,
                    position: 'top-left',
                    duration: 2000,
                    // offset: 100,
                    type: 'error'
                });
            })
            
        }
    },

    mounted(){
        // this.radio = this.zones[0]
        
    }
})
