<template>
  <div id="app">
    <Teams></Teams>
  </div>
</template>

<script>
import Teams from './components/Teams.vue'
import 'bootstrap/dist/css/bootstrap.css';
export default {
  name: "App",
  components: {
    'Teams': Teams
  }
};
</script>
