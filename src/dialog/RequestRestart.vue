<template>
  <v-dialog v-model="showDialog" max-width="500" persistent>
    <v-card>
      <v-list-item>
        <v-card-title class="headline">Restart Game?</v-card-title>
      </v-list-item>
      <v-card-text class="pl-8">{{ dialogText }}</v-card-text>
      <v-card-text class="pl-8">The game will restart once a majority of players are in consensus.</v-card-text>
      <v-card-actions>
        <div class="flex-grow-1"></div>
        <v-btn @click="cancelRequest" color="red" icon>
          <v-icon large>mdi-close-circle-outline</v-icon>
        </v-btn>
        <v-btn @click="sendRequest" color="green" icon>
          <v-icon large>mdi-check-circle-outline</v-icon>
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

const RequestProps = Vue.extend({
  props: {
    showDialog: Boolean,
    fromPlayer: String,
  },
});

@Component({
  components: {},
})
export default class RequestRestart extends RequestProps {

  private get dialogText(): string {
    return (!this.fromPlayer) ?
      'Do you want to request a restart of this game?' :
      `Player ${this.fromPlayer} wants to restart the game. Do you want to?`;
  }

  private sendRequest() {
    this.$emit('request');
  }

  private cancelRequest() {
    this.$emit('cancel');
  }
}
</script>
