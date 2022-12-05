<template>
  <perfect-scrollbar>
    <ul>
      <li>Name: {{this.info.name}}</li>
      <li v-if="this.info.description != ''">Description: {{this.info.description}}</li>
      <li v-if="this.info.path != ''">Path: {{this.info.path}}</li>
      <li>Networks: [
        <ul v-for="network in this.info.networks" :key="network.name" style="margin-top: 0.3em;margin-bottom: 0.3em;">
          <li>Name: {{network.name}}</li>
          <li>CIDR: {{network.cidr}}</li>
        </ul>
        <ul>
          <li>]</li>
        </ul>
      </li>
      <li>PortForwards: [
        <ul v-for="port in this.info.portforwards" :key="(port.sourceport+'_'+port.destinationport+'_'+port.destinationip+'_'+port.protocol+'_'+port.comment)" style="margin-top: 0.3em;margin-bottom: 0.3em;">
          <li>SourcePort: {{port.sourceport}}</li>
          <li v-if="(port.destinationport != '')">DestinationPort: {{port.destinationport}}</li>
          <li>Protocol: {{port.protocol}}</li>
          <li>Destination: {{port.destinationip}}</li>
          <li v-if="port.comment != ''">Comment: {{port.comment}}</li>
        </ul>
        <ul>
          <li>]</li>
        </ul>
      </li>
    </ul>
  </perfect-scrollbar>
</template>

<script>
export default {
  props: ['info'],
}
</script>

<style>

</style>