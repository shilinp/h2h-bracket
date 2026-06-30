import './app.css'
import { mount } from 'svelte'
import UploadApp from './UploadApp.svelte'

const app = mount(UploadApp, {
  target: document.getElementById('app'),
})

export default app
