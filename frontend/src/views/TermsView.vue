<script setup lang="ts">
import LegalPage from '@/components/layout/LegalPage.vue'

const REPO_URL = 'https://github.com/raainshe/Transcendence'
const LICENSE_URL = 'https://github.com/raainshe/Transcendence/blob/main/LICENSE'
</script>

<template>
  <LegalPage title="Terms of Service" updated="May 2026">
    <h2>Acceptance</h2>
    <p>
      By loading or playing this application you agree to these terms. If you do not
      agree, please close the page and refrain from using the project.
    </p>

    <h2>Educational nature</h2>
    <p>
      Transcendence is a non-commercial educational project built as part of the
      <strong>42 ft_transcendence</strong> curriculum. It is provided free of charge for
      learning, demonstration, and personal play. There is no commercial offering, no
      paid tier, and no in-app purchases.
    </p>

    <h2>License</h2>
    <p>
      The source code is published in the project repository and distributed under the
      license that ships with that repository. By using the application you also agree to
      the terms of that license. The current license text can be reviewed at
      <a :href="LICENSE_URL" target="_blank" rel="noopener noreferrer">{{ LICENSE_URL }}</a>.
    </p>

    <h2>Tetris trademark notice</h2>
    <p>
      <strong>"Tetris"</strong> is a registered trademark of Tetris Holding, LLC. This
      project is an unofficial, fan-made educational implementation. It is not produced,
      endorsed, sponsored, or affiliated with Tetris Holding, LLC, The Tetris Company,
      Blue Planet Software, or any of their subsidiaries. All trademarks, gameplay
      mechanics, and references belong to their respective owners and are used here for
      educational and illustrative purposes only.
    </p>

    <h2>No warranty</h2>
    <p>
      The application is provided <strong>"AS IS"</strong>, without warranty of any kind,
      express or implied, including but not limited to the warranties of merchantability,
      fitness for a particular purpose, and non-infringement. The authors make no
      guarantees about availability, correctness, or fitness for any specific use.
    </p>

    <h2>Limitation of liability</h2>
    <p>
      To the maximum extent permitted by applicable law, in no event shall the authors or
      contributors be liable for any direct, indirect, incidental, special, exemplary, or
      consequential damages (including loss of data, loss of profits, or business
      interruption) arising in any way out of the use of, or inability to use, the
      application, even if advised of the possibility of such damages.
    </p>

    <h2>Acceptable use</h2>
    <p>
      You agree to use the application in good faith. Do not attempt to disrupt the
      service, exploit other players (in any future multiplayer modes), or use the
      application to harass others. Be respectful of fellow players and contributors.
    </p>

    <h2>Changes to these terms</h2>
    <p>
      These terms may be updated as the project evolves. Updates are committed to the
      project repository, and the "Last updated" date above reflects the most recent
      revision. Continued use of the application after changes constitutes acceptance of
      the revised terms.
    </p>

    <h2>Contact</h2>
    <p>
      Questions can be raised by opening an issue on the project repository:
      <a :href="REPO_URL" target="_blank" rel="noopener noreferrer">{{ REPO_URL }}</a>.
    </p>
  </LegalPage>
</template>
