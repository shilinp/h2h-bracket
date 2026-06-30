import './app.css'
import { mount } from 'svelte'
import SpecialApp from './SpecialApp.svelte'

const app = mount(SpecialApp, {
  target: document.getElementById('app'),
})

export default app
