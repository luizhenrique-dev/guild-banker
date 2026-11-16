<#import "template.ftl" as native>
<#macro registrationLayout bodyClass="" displayInfo=false displayMessage=true displayRequiredFields=false>
<@native.registrationLayout bodyClass=bodyClass displayInfo=displayInfo displayMessage=displayMessage displayRequiredFields=displayRequiredFields; section>
    <#nested section>
    <#if section = "form">
        <aside class="guild-hero" aria-label="${msg('guildHeroLabel')}">
            <div class="guild-hero-content">
                <h2>${msg("guildHeroTitle")}</h2>
                <p>${msg("guildHeroDescription")}</p>
                <div class="guild-preview" aria-hidden="true">
                    <span>${msg("guildPreviewLabel")}</span>
                    <div class="guild-total guild-balance">
                        <span>${msg("guildBalance")}</span>
                        <strong class="guild-amount">R$ 1.460,00</strong>
                    </div>
                    <div class="guild-totals">
                        <div class="guild-total guild-expense">
                            <div class="guild-total-heading">
                                <span>${msg("guildExpenses")}</span>
                                <svg class="guild-item-icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M12 7v10m-4-4 4 4 4-4"/></svg>
                            </div>
                            <strong class="guild-amount">R$ 3.740,00</strong>
                            <dl class="guild-breakdown">
                                <dt><svg class="guild-item-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M3 3h2l3 12h10l3-9H6"/><circle cx="9" cy="20" r="1"/><circle cx="18" cy="20" r="1"/></svg>${msg("guildMarket")}</dt><dd>R$ 1.240,00</dd>
                                <dt><svg class="guild-item-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="m3 10 9-7 9 7M5 9v12h14V9M9 21v-8h6v8"/></svg>${msg("guildHousing")}</dt><dd>R$ 2.500,00</dd>
                            </dl>
                        </div>
                        <div class="guild-total guild-income">
                            <div class="guild-total-heading">
                                <span>${msg("guildIncome")}</span>
                                <svg class="guild-item-icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M12 17V7m-4 4 4-4 4 4"/></svg>
                            </div>
                            <strong class="guild-amount">R$ 5.200,00</strong>
                            <dl class="guild-breakdown">
                                <dt><svg class="guild-item-icon" viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="7" width="18" height="14" rx="2"/><path d="M8 7V3h8v4M3 12l9 3 9-3M12 12v5"/></svg>${msg("guildSalary")}</dt><dd>R$ 5.200,00</dd>
                            </dl>
                        </div>
                    </div>
                </div>
            </div>
        </aside>
    </#if>
</@native.registrationLayout>
</#macro>
