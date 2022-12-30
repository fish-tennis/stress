
var options = {
    onChange: () => {
        //console.log("on change", this.json.hello)
        dialog.modified = true
    },
    onValidate: function (json) {

    },
    mode: 'text'
}


let dialog = new Vue({
    el: "#dialog",
    data: {
        json: { "hello": "world" },
        options: options,
        title: "test",
        modified:false,
        show: false
    },
    methods: {
        onError(err) {
            console.log("on err", err)
        },
        onClick() {
            //console.log(this.title, this.modified, this.json)
            if (this.modified){
                config.onJsonModified(this.title, this.json)
            }
            this.modified = false
            this.show = false
        },

        Show(title, data) {
            if (data != undefined) {
                this.json = data
            }

            this.title = title
            this.show = true
        }
    }
})