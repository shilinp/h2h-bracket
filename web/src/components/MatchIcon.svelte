<script lang="ts">
    interface Props {
        teamId: number | undefined;
        sizePx: number;
        disabledIcon: boolean
    }

    let { teamId, sizePx, disabledIcon=false }: Props = $props();

    let iconIndex = $derived(teamId != null ? teamId % 4 : null);
</script>

{#if iconIndex === 0}
    <img src= {disabledIcon ? "/sandwich1-disabled.svg" : "/sandwich1.svg"} alt="" width={sizePx} height={sizePx} />
{:else if iconIndex === 1}
    <img src= {disabledIcon ? "/sandwich2-disabled.svg" : "/sandwich2.svg"} alt="" width={sizePx} height={sizePx} />
{:else if iconIndex === 2}
    <img src= {disabledIcon ? "/sandwich3-disabled.svg" : "/sandwich3.svg"} alt="" width={sizePx} height={sizePx} />
{:else if iconIndex === 3}
    <img src= {disabledIcon ? "/sandwich4-disabled.svg" : "/sandwich4.svg"} alt="" width={sizePx} height={sizePx} />
{/if}
