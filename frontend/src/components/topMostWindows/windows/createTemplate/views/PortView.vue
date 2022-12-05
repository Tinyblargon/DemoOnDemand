<template>
  <div class="innerWindowTemplate">
    <div>
      <button @click="this.addLine()">Add</button>
      <button :disabled="(this.selectedLine === 0 || this.portForwards.length < 2 )" @click="this.removeLine(this.selectedLine)">Remove</button>
    </div>
    <table id="pathTable">
      <colgroup>
        <col span="1" id="pathRow0">
        <col span="1" id="pathRow1">
        <col span="1" id="pathRow2">
        <col span="1" id="pathRow3">
        <col span="1">
      </colgroup>
      <tbody>
        <tr>
          <td>
            <label class="pathTableHeader">Source Port</label>
          </td>
          <td>
            <label class="pathTableHeader">Destination IP</label>
          </td>
          <td>
            <label class="pathTableHeader">Destination Port</label>
          </td>
          <td>
            <label class="pathTableHeader">Protocol</label>
          </td>
          <td>
            <label class="pathTableHeader">Comment</label>
          </td>
        </tr>
      </tbody>
    </table>
    <perfect-scrollbar style="height: 15em;">
      <table id="pathTable">
        <colgroup>
          <col span="1" id="pathRow0">
          <col span="1" id="pathRow1">
          <col span="1" id="pathRow2">
          <col span="1" id="pathRow3">
          <col span="1">
        </colgroup>
        <tbody>
          <tr v-for="portForward in this.portForwards" :key="portForward.id" @click="this.setSelectedLine(portForward.id)">
            <td>
              <input type="text" maxlength="5" required v-model="portForward.sourcePort.value" @blur="this.checkSourcePort(portForward)"/>
            </td>
            <td>
              <input type="text" maxlength="39" required v-model="portForward.ip.value" @blur="this.checkIP(portForward)"/>
            </td>
            <td>
              <input type="text" maxlength="5" v-model="portForward.destinationPort.value" @blur="this.checkDestinationPort(portForward)"/>
            </td>
            <td>
              <input type="text" required v-model="portForward.protocol.value" @blur="this.checkProtocol(portForward)"/>
            </td>
            <td>
              <input type="text" v-model="portForward.comment" @blur="this.emit()"/>
            </td>
          </tr>
        </tbody>
      </table>
    </perfect-scrollbar>
    <p v-if="this.error" class="errorText">Invalid format</p>
  </div>
</template>

<script>
import util from './../../../../../util.js'
export default {
  props: ['currentPortForwards'],
  emits: ["state"],
  data() {
    return {
      ready: false,
      portForwards: [
        {
          id: 1,
          sourcePort: {
            value: '',
            error: false
          },
          destinationPort: {
            value: '',
            error: false
          },
          protocol: {
            value: '',
            error: false
          },
          ip: {
            value: '',
            error: false
          },
          comment: '',
          error: false,
        },
      ],
      selectedLine: 0,
      error: false,
    }
  },
  created() {
    if (this.currentPortForwards.length > 0) {
      this.LoadPortForwards()
    }
  },
  methods: {
    LoadPortForwards() {
      this.portForwards = this.currentPortForwards
    },
    setSelectedLine(id) {
      this.selectedLine = id
    },
    addLine() {
      let portForwards = this.portForwards
      portForwards.push({
        id: portForwards[portForwards.length-1].id+1,
        sourcePort: {
            value: '',
            error: false
          },
          destinationPort: {
            value: '',
            error: false
          },
          protocol: {
            value: '',
            error: false
          },
          ip: {
            value: '',
            error: false
          },
          comment: '',
          error: false,
      })
      this.portForwards=portForwards
      this.emit()
    },
    removeLine(id) {
      let arr = this.portForwards
      const objWithIdIndex = arr.findIndex((obj) => obj.id === id)
      if (objWithIdIndex > -1) {
        arr.splice(objWithIdIndex, 1)
      }
      this.portForwards = arr
      this.selectedLine = 0
      this.emit()
    },
    checkSourcePort(row) {
      row.sourcePort.error = util.BoolInvert(this.portValid(row.sourcePort.value))
      this.rowError(row)
    },
    checkIP(row) {
      row.ip.error = util.BoolInvert(this.ipValid(row.ip.value))
      this.rowError(row)
    },
    checkDestinationPort(row) {
      row.sourcePort.error = util.BoolInvert(this.portValid(row.sourcePort.value))
      this.rowError(row)
    },
    checkProtocol(row) {
      row.protocol.error = util.BoolInvert(this.itemInArray(row.protocol.value.toUpperCase(),['','TCP','UDP']))
      this.rowError(row)
    },
    portValid(number) {
      if (number == '' || (number > 0 && number <= 65536 )) {
        return true
      }
      return false
    },
    ipValid(ip) {
      // https://stackoverflow.com/questions/23483855/javascript-regex-to-validate-ipv4-and-ipv6-address-no-hostnames
      var expression = /((^\s*((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))\s*$)|(^\s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?\s*$))/;
      if (ip == '' || expression.test(ip)) {
        return true
      } else {
        return false
      }
    },
    itemInArray(item,array) {
      for (let i = 0; i < array.length; i++) {
        if (item === array[i]) {
          return true
        }
      }
      return false
    },
    rowError(row) {
      if (row.sourcePort.error) {
        row.error = true
      } else if (row.destinationPort.error) {
        row.error = true
      } else if (row.protocol.error) {
        row.error = true
      } else if (row.ip.error) {
        row.error = true
      } else {
        row.error = false
      }
      this.emit()
    },
    globalError() {
      for (let i = 0; i < this.portForwards.length; i++) {
        if (this.portForwards[i].error) {
          this.error = true
          return
        }
      }
      this.error = false
    },
    setReady() {
      if (this.error) {
        this.ready = false
        return
      }
      for (let i = 0; i < this.portForwards.length; i++) {
        if (this.portForwards[i].sourcePort.value == '') {
          this.ready = false
          return
        }
        if (this.portForwards[i].protocol.value == '') {
          this.ready = false
          return
        }
        if (this.portForwards[i].ip.value == '') {
          this.ready = false
          return
        }
      }
      this.ready = true
    },
    emit() {
      this.globalError()
      this.setReady()
      this.$emit('state', {
        info: {
          portForwards: this.portForwards,
        },
        ready: this.ready,
      })
    }
  }
}
</script>

<style>
#pathTable {
  width: 98%;
  /* width: 48em; */
}
#pathRow0 {
  width: 10%;
  /* width: 3.9em; */
}
#pathRow1 {
  width: 34%;
  /* width: 21em; */
}
#pathRow2 {
  width: 13%;
  /* width: 3.9em; */
}
#pathRow3 {
  width: 7%;
  /* width: 3em; */
}
#pathRow4 {
  /* width: 20%; */
  /* width: 18em; */
}
.pathTableHeader {
  margin-top: 0;
  margin-bottom: 0;
}



</style>