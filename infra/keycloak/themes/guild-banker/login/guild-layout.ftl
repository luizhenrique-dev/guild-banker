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
                    <strong class="guild-amount">R$ 8.450,00</strong>
                    <div><span>${msg("guildMarket")}</span><span>R$ 1.240,00</span></div>
                    <div><span>${msg("guildHousing")}</span><span>R$ 2.500,00</span></div>
                    <div><span>${msg("guildIncome")}</span><span>R$ 5.200,00</span></div>
                </div>
            </div>
        </aside>
    </#if>
</@native.registrationLayout>
</#macro>
