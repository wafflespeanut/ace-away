<template>
  <div class="player" v-bind:style="styles"></div>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';
import { TableProps } from './PlayTable.vue';

const START_ANGLE = Math.PI / 2;

// Create extension for hinting TS.
const PlayerProps = Vue.extend({
  props: {
    canMove: Boolean,
    idx: Number,
    numPlayers: Number,
    tableProps: Object as () => TableProps,
  },
});

@Component
export default class Player extends PlayerProps {

  /**
   * @returns Whether this player is the user.
   */
  public get isUser(): boolean {
    return this.idx === 0;
  }

  /**
   * @returns The styles associated with this component.
   */
  private get styles(): object {
    const idx = this.idx;
    const total = this.numPlayers;
    const tProps = this.tableProps;
    if (!this.canMove) {
      return {};
    }

    const angle = 2 * Math.PI / total;
    const x = Math.cos(START_ANGLE + idx * angle) * (tProps.width / 2);
    const y = Math.sin(START_ANGLE + idx * angle) * (tProps.height / 2);
    return {
      transform: `translate(${x + tProps.offsetX}vw, ${y + tProps.offsetY}vh)`,
      backgroundColor: (this.isUser) ? 'black' : '',
    };
  }
}
</script>

<style scoped>
.player {
  position: absolute;
  top: 20vh;
  left: 46vw;
  width: 8vw;
  height: 8vw;
  border-radius: 50%;
  border: 2px solid;
  transition: all 250ms ease;
}

@media screen and (min-width: 700px) {
  .player {
    top: 30vh;
    left: 48vw;
    width: 4vw;
    height: 4vw;
  }
}
</style>
