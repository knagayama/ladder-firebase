<template>
  <ol>
    <li v-for="team in teams" v-bind:key="team.name">{{ team.name }}: {{team.player1}}, {{ team.player2 }}, {{ team.player3 }}, {{ team.player4 }}, {{ team.player5 }},  </li>
  </ol>
</template>

<script>
import firebase from "firebase";
export default {
  name: "Teams",
  data() {
    return {
      teams: []
    };
  },
  mounted: function() {
    var db = firebase.firestore();
    db.collection("tournaments/spladder4/teams")
      .get()
      .then(querySnapshot => {
        querySnapshot.forEach(doc => {
          console.log(doc.data());
          this.teams.push(doc.data());
        });
      });
  }
};
</script>
