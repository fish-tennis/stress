const CodeSuccess = 0;
const CodeFailed = 1;


class Agent {
    constructor(GO) {
        this.reg = {}
        this.seq = 0
        this.GO = GO
    }

    Invoke(call) {
        let fn = this.obt(call.method);
        let callRst = {
            seq: call.seq,
        };
        if (!fn) {
            callRst.code = CodeFailed;
            callRst.msg = "unknow action " + call.method;
            console.log("动作未注册", method)
        } else {
            callRst.code = CodeSuccess;
            callRst.msg = call.msg;
            callRst.data = fn(call.arg);
            if (callRst.data !== true) callRst.code = CodeFailed
        }
        if (callRst.data !== false) {
            // console.log("[GO." + call.seq + "]调用", call);
            // console.log("[GO." + call.seq + "]响应", callRst);
        }
        return callRst;
    }

    Reg(action, fn) {
        if (this.reg[action]) {
            console.error("动作已注册", action)
        }
        this.reg[action] = fn
    }

    obt(act) {
        return this.reg[act]
    }

    Call(act, arg, msg = "") {
        this.seq++
        let call = {
            seq: this.seq,
            method: act,
            msg: msg,
            arg: arg,
        };
        // console.log("[JS." + call.seq + "]调用", call);
        return this.GO(call).then(callRst => {
            // console.log("[JS." + call.seq + "]响应", callRst);
            if (callRst.seq !== call.seq) {
                return new Promise(function (resolve, reject) {
                    reject("seq 不相等")
                })
            }
            if (callRst.code !== CodeSuccess) {
                return new Promise(function (resolve, reject) {
                    reject("请求失败" + callRst.msg);
                })
            }
            return callRst.data
        })
    }
}


if (!GO) {
    window.alert("请在客户端环境下运行")
}
const GOAgent = new Agent(GO)

class Lock {
    constructor() {
        this.ts = 0
    }

    GetTime() {
        return Math.floor(new Date().getTime() / 1000)
    }

    isLock() {
        return this.ts > this.GetTime();
    }

    Lock(expire = 5) {
        if (this.isLock()) {
            return false
        }
        this.ts = this.GetTime() + expire;
        return true
    }

    Unlock() {
        if (!this.isLock()) {
            return false
        }
        this.ts = 0;
        return true
    }
}


class FileAgent {
    constructor(filename, name = "") {
        this.lock = new Lock()
        this.filename = filename
        this.name = name
        this.cache = {}
        return new Promise( (resolve, reject) => {
            return GOAgent.Call("ReadFile", this.filename, "读取文件").then(buf => {
                this.cache = this.parse(buf);
                resolve(this)
            }).catch(e => {
                // console.log("创建文件：", this.filename)
                // this.create = true
                // return this.Sets({}).then(data => {
                //     resolve(this)
                // })
            })
        })
    }

    parse(str) {
        let rs = {};
        try {
            rs = JSON.parse(str)
        } catch (e) {
            return rs
        }
        return rs
    }

    string(obj) {
        return JSON.stringify(obj, "", "    ")
    }

    Gets() {
        return this.cache
    }

    Get(idx, dft = "") {
        return this.cache[idx] === undefined ? dft : this.cache[idx]
    }

    // 覆盖整个文件
    Sets(obj) {
        if (this.lock.Lock()) {
            this.cache = obj
            return GOAgent.Call("WriteFile", [this.filename, this.string(obj)], "写入文件").then(rst => {
                // console.log('WriteFile', rst, obj)
                this.lock.Unlock();
                return rst;
            })
        } else {
            console.error("获取写锁失败", this.filename, obj, this.lock.GetTime(), this.lock.isLock());
            return new Promise(function (resolve, reject) {
                resolve(false)
            })
        }
    }

    Find(data, key) {
        
        let keys = key.split('.')
        let prefix_keys = keys.slice(0, keys.length-1)
        key = keys[keys.length-1]
        // console.log("find", keys, prefix_keys)
        let curr = data
        for (const k of prefix_keys){
            if (!curr.hasOwnProperty(k)){
                curr[k] = {}
            }
            curr = curr[k]
        }
        return [curr, key]
    }

    // 增量写
    Set(key, val) {
        let data = this.Gets()
        let [curr, k] = this.Find(data, key)
        // console.log("set",curr, k)
        curr[k] = val
        return this.Sets(data)
    }

    Del(key) {
        let data = this.Gets()
        let [curr, k] = this.Find(data, key)
        delete curr[k]
        return this.Sets(data)
    }
}


class RequestAgent {

    constructor(server) {
        this.name = server
    }

    Gets() {
        return GOAgent.Call("GetMsgList", "", "获取cs消息列表")
    }

    GetInfo(req) {
        return GOAgent.Call("GetMsgDetail", req, "获取单个请求详情")
    }

    Stop() {
        return GOAgent.Call("Stop", "", "stopApp")
    }

    Reload() {
        return GOAgent.Call("ReloadConfig", "", "reload config")
    }



    Send(req, info) {
        return GOAgent.Call("SendReq", [req, info], "发送请求")
    }
}


class RobotAgent {
    constructor() {    }

    RunBehaviorTree(treeID, tab, data) {
        return GOAgent.Call("StartRobot", [treeID, tab, data], "运行行为树")
    }
    StressBehaviorTree(serverConfig,  sinfo) {
        return GOAgent.Call("StressRobot", [serverConfig,  sinfo], "压测行为树")
    }
}
