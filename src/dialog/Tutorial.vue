<template>
  <v-dialog v-bind:value="showDialog" max-width="550" persistent>
    <v-card>
      <v-list-item>
        <v-card-title class="headline">Tour of Ace Away!</v-card-title>
        <v-row class="pr-1" align="center" justify="end">
          <v-btn @click="reset" color="red" icon>
            <v-icon large>mdi-close-circle-outline</v-icon>
          </v-btn>
        </v-row>
      </v-list-item>
      <v-card-text v-for="(msg, i) in steps[stepi].msgs" :key="i"
                   class="pl-8">{{ msg }}</v-card-text>
      <v-card-actions>
        <div class="flex-grow-1"></div>
        <v-btn :disabled="stepi <= 1" @click="stepBack" color="green" icon>
          <v-icon large>mdi-chevron-left-circle-outline</v-icon>
        </v-btn>
        <v-btn :disabled="stepi == steps.length - 1" @click="stepForward" color="green" icon>
          <v-icon large>mdi-chevron-right-circle-outline</v-icon>
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

interface TutorialStep {
  msgs?: string[];
  room?: string | null;
}

const TutorialProps = Vue.extend({
  props: {
    playerNeedsHelp: Boolean,
  },
});

@Component({
  components: {},
})
export default class Tutorial extends TutorialProps {

  private get showDialog(): boolean {
    return this.playerNeedsHelp && this.steps[this.stepi].msgs !== undefined;
  }

  private set showDialog(v: boolean) {
    // We don't need to set anything, because the dialog is persistent
    // and when it gets closed, we notify the parent anyway.
  }

  private stepi: number = 0;

  private steps: TutorialStep[] = [{
    msgs: [
      `This takes you on a tour on how to play this game.
 You can skip this and start playing whenever you feel like.`,
    ],
  }, {
    msgs: [
      `Ace is a game of cards.`,
      `For a single deck of cards, 4-6 players can play at a time. Although,
 there's nothing stopping 2 or even 10 players from playing - only that the game
 wouldn't be as interesting.`,
      `The goal is to get rid of all cards in
 your hand as soon as possible and "escape" the room.
 The player with leftover cards loses the game.`,
    ],
    room: 'test',
  }, {
    msgs: [
      `The game begins with cards distributed across a set number of players.`,
      `For the purpose of this tutorial, let's start a game with 4 players.`,
    ],
  }];

  private reset() {
    this.stepi = 0;
    this.$emit('close');
  }

  private stepBack() {
    this.stepi -= 1;
    this.handleStep();
  }

  private stepForward() {
    this.stepi += 1;
    this.handleStep();
  }

  private handleStep() {
    const stepInfo = this.steps[this.stepi];
    this.$emit('tutorial-step', stepInfo);
    if (!stepInfo.msgs) {
      this.showDialog = false;
    }
  }
}

export { TutorialStep };
</script>
