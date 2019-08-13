import Vue from 'vue';
import VModal from 'vue-js-modal';
import App from './App.vue';
import vSelect from 'vue-select';
import 'vue-select/dist/vue-select.css';

Vue.config.productionTip = false;
Vue.component('v-select', vSelect);
Vue.use(VModal, { dialog: true });

new Vue({
  render: (h) => h(App),
}).$mount('#app');
