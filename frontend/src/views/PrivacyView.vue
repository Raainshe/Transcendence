<script setup lang="ts">
import LegalPage from '@/components/layout/LegalPage.vue'

const REPO_URL = 'https://github.com/raainshe/Transcendence'
</script>

<template>
  <LegalPage title="Privacy Policy" updated="May 2026">
    <h2>Overview</h2>
    <p>
      Transcendence is an educational project built for the
      <strong>42 ft_transcendence</strong> curriculum. It is a fan-made Tetris-style game
      maintained by a small group of students. We do not run analytics, advertising, or
      tracking of any kind. This page describes the very limited data the application
      handles so you know exactly what happens in your browser when you play.
    </p>

    <h2>Information we do not collect</h2>
    <ul>
      <li>No accounts, names, email addresses, or other personal identifiers.</li>
      <li>No analytics, telemetry, or usage tracking.</li>
      <li>No advertising or third-party tracking cookies.</li>
      <li>No fingerprinting or behavioural profiling.</li>
    </ul>

    <h2>Local browser storage</h2>
    <p>
      Your gameplay preferences (such as the selected variation and player count) are kept
      in a Pinia store inside the running tab. Where the project persists those
      preferences, it does so using your browser's own storage (for example
      <code>localStorage</code>). This data never leaves your device, is never sent to a
      server, and can be cleared at any time from your browser's site-data controls.
    </p>

    <h2>Third-party content</h2>
    <p>
      The Credits page loads developer avatars directly from GitHub
      (<code>github.com/&lt;handle&gt;.png</code>). When you visit that page, your browser
      makes a request to GitHub to fetch the image. Those requests are handled by GitHub
      under their own privacy policy. No other third-party services are contacted by the
      application.
    </p>

    <h2>Children's privacy</h2>
    <p>
      Because we do not collect any personal data, the application does not knowingly
      gather information from anyone, including children. There is no account system and
      no age gate.
    </p>

    <h2>Your controls</h2>
    <p>
      You can clear any locally stored preferences at any time by clearing site data for
      this domain from your browser's settings. Doing so will reset gameplay preferences
      to their defaults the next time you load the page.
    </p>

    <h2>Changes to this policy</h2>
    <p>
      As the project evolves, this policy may change. Updates will be committed to the
      project repository, and the "Last updated" date above will reflect the most recent
      revision.
    </p>

    <h2>Contact</h2>
    <p>
      Questions or concerns can be raised by opening an issue on the project repository:
      <a :href="REPO_URL" target="_blank" rel="noopener noreferrer">{{ REPO_URL }}</a>.
    </p>
  </LegalPage>
</template>
