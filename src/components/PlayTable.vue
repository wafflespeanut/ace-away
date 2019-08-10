<template>
  <div>
    <div id="banner">{{ bannerMsg }}</div>
    <div id="message">{{ msg }}</div>
    <div id="playerSelection">
      <span>No. of players: </span>
      <span class="select">
        <v-select :options="[3, 4, 5, 6]" @input="playersSelected" placeholder="Choose..." />
      </span>
    </div>
    <div id="cardSelection" v-if="hasPlayers">
      <div>
        <span class="suite"
              v-for="(suite, i) in suites"
              v-bind:key="i"
              v-bind:class="{ bright: suite.bright, selected: selectionMatches(suite) }"
              @click="selectedSuite = selectionMatches(suite) ? null : suite">{{ suite.display }}</span>
      </div>
      <div>
        <span class="label"
              v-for="(label, i) in labels"
              v-bind:key="i"
              v-bind:class="{ bright: (selectedSuite != null && selectedSuite.bright), selected: selectedLabel == label }"
              @click="selectedLabel = (selectedLabel == label) ? null : label">{{ label }}</span>
      </div>
    </div>
    <div id="tableArea">
      <Player v-for="(p, idx) in players"
              v-bind:key="idx"
              v-bind:idx="idx"
              v-bind:numPlayers="p.total"
              v-bind:canMove="p.moveable"
              v-bind:tableProps="tableProps" />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

import Player from './Player.vue';

/**
 * Removes the element matching the given selector by zero'ing its
 * opacity and removing itself from its parent.
 *
 * @param selector The query selector used for finding the element.
 * @param timeout Timeout for the element to smoothly disappear
 * from the screen (should match the CSS transition).
 */
function removeElementWithSelector(selector: string, timeout: number) {
  const elem: HTMLElement | null = document.querySelector(selector);
  if (elem === null) {
    return;
  }

  elem.style.opacity = '0';
  setTimeout(() => {
    if (elem.parentElement) {
      elem.parentElement.removeChild(elem);
    }
  }, timeout);
}

/**
 * Timeout (ms) for some actions (fading in/out).
 */
const ACTION_TIMEOUT = 1000;

const SUITES = [{
  name: 'diamond',
  short: 'd',
  display: '♦',
  bright: true,
}, {
  name: 'clover',
  short: 'c',
  display: '♣',
  bright: false,
}, {
  name: 'heart',
  short: 'h',
  display: '♥',
  bright: true,
}, {
  name: 'spade',
  short: 's',
  display: '♠',
  bright: false,
}];

interface Suite {
  name: string;
  short: string;
  display: string;
  bright: boolean;
}

interface TableProperties {
  width: number;
  height: number;
  offsetX: number;
  offsetY: number;
}

@Component({
  components: {
    Player,
  },
})
export default class PlayTable extends Vue {

  private players: object[] = [];

  private msg: string = '';

  private bannerMsg: string = '';

  private readonly suites: Suite[] = SUITES;

  private readonly labels: string[] = ['2', '3', '4', '5', '6', '7', '8', '9', '10', 'J', 'Q', 'K', 'A'];

  private selectedSuite: Suite | null = null;

  private selectedLabel: string | null = null;

  /**
   * @returns Some properties of the table in viewport width units (vw).
   */
  private get tableProps(): TableProperties {
    const offsetX = 0;
    const width = 65;
    let height = 40;
    let offsetY = 3;
    if (screen.width > 700) {
      height = 40;
      offsetY = 8;
    }

    return {
      width,
      height,
      offsetX,
      offsetY,
    };
  }

  private get hasPlayers(): boolean {
    return this.players.length > 0;
  }

  public displayMessage(banner: string, msg: string) {
    let msgEl = this.$el.querySelector('#message');
    let banEl = this.$el.querySelector('#banner');
    msgEl.style.opacity = '0';
    banEl.style.opacity = '0';

    setTimeout(() => {
      this.msg = msg;
      this.bannerMsg = banner;
      msgEl.style.opacity = '1';
      banEl.style.opacity = '1';
    }, ACTION_TIMEOUT);
  }

  /**
   * Triggered once the user has selected the no. of players.
   * This adds child (`Player`) components.
   *
   * @param total No. of players in this table.
   */
  private playersSelected(total: number) {
    console.debug(`Players selected: ${total}`);
    removeElementWithSelector('#playerSelection', ACTION_TIMEOUT);
    this.displayMessage('Add your cards.', 'Pick a suite and a label to add your card.');

    setTimeout(() => {
      for (const idx of Array(total).keys()) {
        this.players.push({ total, moveable: false });
        setTimeout(() => {
          this.players[idx].moveable = true;
        }, (idx + 1) * 250);
      }
    }, ACTION_TIMEOUT);
  }

  private selectionMatches(suite: Suite): boolean {
    return this.selectedSuite != null && this.selectedSuite.short == suite.short
  }
}

export { TableProperties };
</script>

<style scoped>
#banner {
  font-size: 3.5vh;
}

#message, #banner {
  opacity: 0;
}

#message, #banner, #playerSelection {
  transition: opacity ease 1s;
}

#playerSelection {
  display: flex;
  align-items: center;
  justify-content: center;
}

#playerSelection > span {
  margin: 5px;
}

#playerSelection > .select {
  width: 60vw;
}

#cardSelection > div {
  margin: 0 10px;
  display: flex;
  flex-flow: wrap;
  align-items: center;
  justify-content: center;
}

#cardSelection .label {
  margin: 0 5px;
  padding: 5px;
  font-size: 4vh;
}

#cardSelection .suite {
  font-size: 9vh;
  padding: 0 5px;
}

#cardSelection .suite, #cardSelection .label {
  border-radius: 25%;
}

.suite.bright, .label.bright {
  color: red;
}

.suite.selected, .label.selected {
  background-color: black;
  color: white;
}

.suite.selected.bright, .label.selected.bright {
  background-color: red;
}

@media screen and (min-width: 700px) {
  #banner {
    font-size: 2.75vh;
  }

  #playerSelection > .select {
    width: 27vw;
  }
}
</style>
