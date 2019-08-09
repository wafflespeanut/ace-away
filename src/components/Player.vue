<template>
  <div class="player" v-bind:style="styles"></div>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

const START_ANGLE = Math.PI / 2;

@Component({
  props: {
    canMove: Boolean,
    idx: Number,
    numPlayers: Number,
    tableSize: Number,
  },
})
export default class Player extends Vue {

  /**
   * @returns Whether this player is the user.
   */
  public get isUser(): boolean {
    return this['idx'] === 0;
  }

  /**
   * @returns The styles associated with this component.
   */
  private get styles(): object {
    const idx = this['idx'];
    const total = this['numPlayers'];
    const tSize = this['tableSize'];
    if (!this['canMove']) {
      return {};
    }

    const angle = 2 * Math.PI / total;
    const x = Math.cos(START_ANGLE + idx * angle) * (tSize / 2);
    const y = Math.sin(START_ANGLE + idx * angle) * (tSize / 2);
    return {
      transform: `translate(${x}vw, ${y}vw)`,
      backgroundColor: (this.isUser) ? 'black' : '',
    };
  }
}
</script>

<style scoped>
.player {
  position: absolute;
  top: 35vh;
  left: 45vw;
  width: 8vw;
  height: 8vw;
  border-radius: 50%;
  border: 2px solid;
  transition: all 250ms ease;
}

@media screen and (min-width: 700px) {
  .player {
    top: 40vh;
    left: 46vw;
    width: 4vw;
    height: 4vw;
  }
}
</style>
