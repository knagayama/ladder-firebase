import Vue from "vue";
import App from "./App.vue";
import firebase from "firebase";
import { BootstrapVue, IconsPlugin } from "bootstrap-vue";

Vue.config.productionTip = false;

var config = {
  apiKey: "AIzaSyDLLfIPe0ccWs2v0USU7az93TSksxYT8xo",
  authDomain: "splathon-ladder.firebaseapp.com",
  databaseURL: "https://splathon-ladder.firebaseio.com",
  projectId: "splathon-ladder",
  storageBucket: "splathon-ladder.appspot.com",
  messagingSenderId: "473538997526",
  appId: "1:473538997526:web:db7cc8ec78e46dd3cbf437",
};
firebase.initializeApp(config);

new Vue({
  render: (h) => h(App),
}).$mount("#app");

Vue.use(BootstrapVue);
Vue.use(IconsPlugin);
