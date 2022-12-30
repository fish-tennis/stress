
let btree = new Vue({
    el: "#btree",
    data: {
        options: options,
        title: "行为树编辑",
        modified: false,
        show: false
    },
    methods: {
        Show(){
            this.show = true
        },

        onClose(){
            config.reloadTree()
            // console.log('onclose')
        }

    }
})