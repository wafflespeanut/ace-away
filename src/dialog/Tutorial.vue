<template>
  <v-dialog v-model="showDialog" max-width="550" persistent>
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
        <v-btn v-show="stepi > 0" @click="stepBack" color="green" icon>
          <v-icon large>mdi-chevron-left-circle-outline</v-icon>
        </v-btn>
        <v-btn v-show="stepi < steps.length - 1" @click="stepForward" color="green" icon>
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
  msgs: string[];
}

const TutorialProps = Vue.extend({
  props: {
    showDialog: Boolean,
  },
});

@Component({
  components: {},
})
export default class Tutorial extends TutorialProps {

  private stepi: number = 0;

  private steps: TutorialStep[] = [{
    msgs: [
      'This takes you through a tour on how to play this game. You can skip this tutorial whenever you feel like.',
    ],
  }];

  private reset() {
    this.stepi = 0;
    this.$emit('close');
  }

  private stepBack() {
    this.stepi -= 1;
  }

  private stepForward() {
    this.stepi += 1;
  }
}
</script>
