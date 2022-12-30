var ss = `
<div v-if="log.show" v-bind:class="log.cls" v-bind:id="log.id">
<div class="button-group fill op">
  <button class="button cate">{{log.cate}}</button>
  <button class="button id" @click=view(log)>{{log.name}}</button>
  <button class="button"@click=ondel(log)><i class="fa fa-trash-o"></i></button>
</div>
<div type="text" class="content" readonly="readonly" v-html=log.txt>
</div>
</div>
`



var renderJson = function (obj) {
    // let obj = (typeof str == "string") ? this.$.parseJSON(str) : str
    let json = JSON.stringify(obj, undefined, 3);
    let reg = /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g
    json = json.replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
    return json.replace(reg, (match) => {
        let cls = 'number';
        if (/^"/.test(match)) {
            cls = /:$/.test(match) ? "key" : "string";
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '"' + '>' + match + "</span>";
    });
}


Vue.component('logshow', {
    props: ['log'],
    methods: {
        ondel: function (e) {
            // console.log("ondel", e)
            e.show = false
            config.saveLogFilter(e.name)
        },
        view(e){
            logdialog.Show(e)
            // console.log("view", e)
        }
    },
    template: ss
})

let logbox = new Vue({
    el: "#logbox",
    data: {
        logs: [],
        tabs:[
            {title:"default", name:1, Content: "hello"}
        ],
    },
    methods: {
        appendlog(log) {
            if (config.skipShowLog(log)){
                return
            }
            let meta = log.meta
            meta.id = `log-${meta.id}`
            meta.cls = `log-item hall ${meta.name} ${meta.type}`
            meta.cate = "logic"
            meta.txt = renderJson(log.data)
            meta.data = log.data
            meta.show = true

            let tab = tabs.current()
            if (meta.tab != 0) {
                 tab = tabs.getById(meta.tab)
            }
            tab.logs.unshift(meta)
        },
        clearAll() {
            let tab = tabs.current()
            tab.logs = []
            this.logs = tab.logs
        },
        onEdit(name, action) {
            console.log(name, action)
        }, 
        setData(logs){
            this.logs = logs
        },

    },
    mounted() {
        GOAgent.Reg('info.log', this.appendlog)
        // this.appendlog({ id: "1", name: "apple" })
        setTimeout(() => {
        }, 3000);
    }
})

let logdialog = new Vue({
    el: "#logdialog",
    data: {
        show: false,
        content:"",
        json:{},
        options:{
            mode: 'view',
            mainMenuBar: true,
            statusBar: false,
            navigationBar: false,
            modes: ['text', 'code', 'view'],
            onEditable(){
                return false
            },
        },
        title:""
    },
    methods: {
        Show(item){
            this.title = item.name
            this.json = item.data
            this.show = true
        },
        onError(){

        }
    }
})
