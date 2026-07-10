<script lang="ts">
    interface Props {
        teamId: number | undefined;
        sizePx: number;
    }

    let { teamId, sizePx }: Props = $props();

    let iconIndex = $derived(teamId != null ? teamId % 4 : null);
</script>

{#if iconIndex === 0}
    <img src="/sandwich1.svg" alt="" width={sizePx} height={sizePx} />
{:else if iconIndex === 1}
    <img src="/sandwich2.svg" alt="" width={sizePx} height={sizePx} />
{:else if iconIndex === 2}
    <img src="/sandwich3.svg" alt="" width={sizePx} height={sizePx} />
{:else if iconIndex === 3}
    <img src="/sandwich4.svg" alt="" width={sizePx} height={sizePx} />
{/if}
