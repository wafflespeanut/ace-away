<template>
  <v-app id="inspire">
    <v-navigation-drawer v-model="drawerOpen" app>
      <!--  -->
    </v-navigation-drawer>

    <v-app-bar app>
      <v-app-bar-nav-icon @click.stop="drawerOpen = !drawerOpen"></v-app-bar-nav-icon>
      <v-toolbar-title>Ace Away!</v-toolbar-title>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <v-alert v-if="alertMsg" border="top" colored-border type="info" elevation="2">{{ alertMsg }}</v-alert>
        <v-item-group mandatory>
          <v-row>
            <v-col class="d-flex justify-center" v-for="(card, i) in hand" :key="i">
              <v-item v-slot:default="{ active, toggle }">
                <v-slide-x-transition>
                  <v-card :color="active ? ( card.suite == 'h' || card.suite == 'd' ? 'red' : 'blue' ) : ''"
                          class="d-flex align-center justify-center"
                          height="100"
                          width="100"
                          @click="toggle">
                    <span :class="active ? 'display-1 flex-grow-1 text-center' : ''">{{ card.label }} {{ prettyMap[card.suite] }}</span>
                  </v-card>
                </v-slide-x-transition>
              </v-item>
            </v-col>
          </v-row>
        </v-item-group>
      </v-container>
      <JoinRoom @joined="playerJoined" :players="allowedPlayers" :showDialog="!roomJoined" />
    </v-content>
    <v-snackbar :value="true" :timeout="0" v-for="(text, i) in notifications" :key="i">
      {{ text }}
      <v-btn color="pink" text @click="notifications.splice(i, 1)">Close</v-btn>
    </v-snackbar>
    <v-footer app>
      <span class="white--text">&copy; 2019 @wafflespeanut</span>
    </v-footer>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

import JoinRoom from './dialog/JoinRoom.vue';
import { Card, Suite, suitePrettyMap, Label } from './deck';
import { ClientMessage, RoomCreationRequest, ServerMessage, RoomResponse } from './persistence/model';
import ConnectionProvider from './persistence/connection';

const ALLOWED_PLAYERS: number[] = [3, 4, 5, 6];

@Component({
  components: {
    JoinRoom,
  },
})
export default class App extends Vue {

  /* Constants used by models */

  private allowedPlayers: number[] = ALLOWED_PLAYERS;

  private prettyMap: any = suitePrettyMap;

  /* Models */

  private drawerOpen: boolean = false;

  private roomJoined: boolean = true;

  private notifications: string[] = [];

  private alertMsg: string | null = null;

  private hand: Card[] = [{
    label: Label.Two,
    suite: Suite.Diamond,
  }, {
    label: Label.Six,
    suite: Suite.Clover,
  }, {
    label: Label.King,
    suite: Suite.Spade,
  }, {
    label: Label.Ace,
    suite: Suite.Heart,
  }, {
    label: Label.Ace,
    suite: Suite.Spade,
  }, {
    label: Label.Six,
    suite: Suite.Heart,
  }, {
    label: Label.King,
    suite: Suite.Clover,
  }];

  /* Internal properties */

  private conn = new ConnectionProvider();

  private created() {
    this.conn.onError(this.showError, true);
  }

  private playerJoined(player: string, resp: ServerMessage<RoomResponse>) {
    if (resp.player === player) {
      this.roomJoined = true;
    }

    const diff = resp.response.max - resp.response.players.length;
    if (diff > 0) {
      this.alertMsg = `Waiting for ${diff} more player(s).`;
    } else {
      this.alertMsg = null;
    }
  }

  private showError(msg: string) {
    this.notifications.push(msg);
  }
}
export { ALLOWED_PLAYERS };
</script>
