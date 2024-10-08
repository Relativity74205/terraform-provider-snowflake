package resources_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	acc "github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/acceptance"
	r "github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/resources"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/internal/provider"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/provider/resources"
	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAcc_DatabaseWithUnderscore(t *testing.T) {
	prefix := acc.TestClient().Ids.AlphaWithPrefix("_")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(tfversion.Version1_5_0),
		},
		CheckDestroy: acc.CheckDestroy(t, resources.DatabaseOld),
		Steps: []resource.TestStep{
			{
				Config: dbConfig(prefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.db", "name", prefix),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "comment", "test comment"),
					resource.TestCheckResourceAttrSet("snowflake_database_old.db", "data_retention_time_in_days"),
				),
			},
		},
	})
}

func TestAcc_Database(t *testing.T) {
	prefix := acc.TestClient().Ids.Alpha()
	prefix2 := acc.TestClient().Ids.Alpha()

	secondaryAccountName := acc.SecondaryTestClient().Context.CurrentAccount(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(tfversion.Version1_5_0),
		},
		CheckDestroy: acc.CheckDestroy(t, resources.DatabaseOld),
		Steps: []resource.TestStep{
			{
				Config: dbConfig(prefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.db", "name", prefix),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "comment", "test comment"),
					resource.TestCheckResourceAttrSet("snowflake_database_old.db", "data_retention_time_in_days"),
				),
			},
			// RENAME
			{
				Config: dbConfig(prefix2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.db", "name", prefix2),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "comment", "test comment"),
					resource.TestCheckResourceAttrSet("snowflake_database_old.db", "data_retention_time_in_days"),
				),
			},
			// CHANGE PROPERTIES
			{
				Config: dbConfig2(prefix2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.db", "name", prefix2),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "comment", "test comment 2"),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "data_retention_time_in_days", "3"),
				),
			},
			// ADD REPLICATION
			// proves https://github.com/Snowflake-Labs/terraform-provider-snowflake/issues/2369 error
			{
				Config: dbConfigWithReplication(prefix2, secondaryAccountName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.db", "name", prefix2),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "comment", "test comment 2"),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "data_retention_time_in_days", "3"),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "replication_configuration.0.accounts.#", "1"),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "replication_configuration.0.accounts.0", secondaryAccountName),
				),
			},
			// IMPORT
			{
				ResourceName:            "snowflake_database_old.db",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replication_configuration"},
			},
		},
	})
}

func TestAcc_DatabaseRemovedOutsideOfTerraform(t *testing.T) {
	id := acc.TestClient().Ids.RandomAccountObjectIdentifier()
	name := id.Name()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(tfversion.Version1_5_0),
		},
		CheckDestroy: acc.CheckDestroy(t, resources.DatabaseOld),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: map[string]config.Variable{
					"db": config.StringVariable(name),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.db", "name", name),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "comment", "test comment"),
					testAccCheckDatabaseExistence(t, id, true),
				),
			},
			{
				PreConfig:       func() { acc.TestClient().Database.DropDatabaseFunc(t, id)() },
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: map[string]config.Variable{
					"db": config.StringVariable(name),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.db", "name", name),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "comment", "test comment"),
					testAccCheckDatabaseExistence(t, id, true),
				),
			},
		},
	})
}

// proves https://github.com/Snowflake-Labs/terraform-provider-snowflake/issues/2021
func TestAcc_Database_issue2021(t *testing.T) {
	name := acc.TestClient().Ids.Alpha()

	secondaryAccountName := acc.SecondaryTestClient().Context.CurrentAccount(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(tfversion.Version1_5_0),
		},
		CheckDestroy: acc.CheckDestroy(t, resources.DatabaseOld),
		Steps: []resource.TestStep{
			{
				Config: dbConfigWithReplication(name, secondaryAccountName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.db", "name", name),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "replication_configuration.0.accounts.#", "1"),
					resource.TestCheckResourceAttr("snowflake_database_old.db", "replication_configuration.0.accounts.0", secondaryAccountName),
					testAccCheckIfDatabaseIsReplicated(t, name),
				),
			},
		},
	})
}

// proves https://github.com/Snowflake-Labs/terraform-provider-snowflake/issues/2356 issue is fixed.
func TestAcc_Database_DefaultDataRetentionTime(t *testing.T) {
	id := acc.TestClient().Ids.RandomAccountObjectIdentifier()

	configVariablesWithoutDatabaseDataRetentionTime := func() config.Variables {
		return config.Variables{
			"database": config.StringVariable(id.Name()),
		}
	}

	configVariablesWithDatabaseDataRetentionTime := func(databaseDataRetentionTime int) config.Variables {
		vars := configVariablesWithoutDatabaseDataRetentionTime()
		vars["database_data_retention_time"] = config.IntegerVariable(databaseDataRetentionTime)
		return vars
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(tfversion.Version1_5_0),
		},
		CheckDestroy: acc.CheckDestroy(t, resources.DatabaseOld),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					revertParameter := acc.TestClient().Parameter.UpdateAccountParameterTemporarily(t, sdk.AccountParameterDataRetentionTimeInDays, "5")
					t.Cleanup(revertParameter)
				},
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithoutDataRetentionSet"),
				ConfigVariables: configVariablesWithoutDatabaseDataRetentionTime(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", r.IntDefaultString),
					checkAccountAndDatabaseDataRetentionTime(t, id, 5, 5),
				),
			},
			{
				PreConfig: func() {
					_ = acc.TestClient().Parameter.UpdateAccountParameterTemporarily(t, sdk.AccountParameterDataRetentionTimeInDays, "10")
				},
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithoutDataRetentionSet"),
				ConfigVariables: configVariablesWithoutDatabaseDataRetentionTime(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", r.IntDefaultString),
					checkAccountAndDatabaseDataRetentionTime(t, id, 10, 10),
				),
			},
			{
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithDataRetentionSet"),
				ConfigVariables: configVariablesWithDatabaseDataRetentionTime(5),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", "5"),
					checkAccountAndDatabaseDataRetentionTime(t, id, 10, 5),
				),
			},
			{
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithDataRetentionSet"),
				ConfigVariables: configVariablesWithDatabaseDataRetentionTime(15),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", "15"),
					checkAccountAndDatabaseDataRetentionTime(t, id, 10, 15),
				),
			},
			{
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithoutDataRetentionSet"),
				ConfigVariables: configVariablesWithoutDatabaseDataRetentionTime(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", r.IntDefaultString),
					checkAccountAndDatabaseDataRetentionTime(t, id, 10, 10),
				),
			},
			{
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithDataRetentionSet"),
				ConfigVariables: configVariablesWithDatabaseDataRetentionTime(0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", "0"),
					checkAccountAndDatabaseDataRetentionTime(t, id, 10, 0),
				),
			},
			{
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithDataRetentionSet"),
				ConfigVariables: configVariablesWithDatabaseDataRetentionTime(3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", "3"),
					checkAccountAndDatabaseDataRetentionTime(t, id, 10, 3),
				),
			},
		},
	})
}

// proves https://github.com/Snowflake-Labs/terraform-provider-snowflake/issues/2356 issue is fixed.
func TestAcc_Database_DefaultDataRetentionTime_SetOutsideOfTerraform(t *testing.T) {
	id := acc.TestClient().Ids.RandomAccountObjectIdentifier()

	configVariablesWithoutDatabaseDataRetentionTime := func() config.Variables {
		return config.Variables{
			"database": config.StringVariable(id.Name()),
		}
	}

	configVariablesWithDatabaseDataRetentionTime := func(databaseDataRetentionTime int) config.Variables {
		vars := configVariablesWithoutDatabaseDataRetentionTime()
		vars["database_data_retention_time"] = config.IntegerVariable(databaseDataRetentionTime)
		return vars
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { acc.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(tfversion.Version1_5_0),
		},
		CheckDestroy: acc.CheckDestroy(t, resources.DatabaseOld),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					revertParameter := acc.TestClient().Parameter.UpdateAccountParameterTemporarily(t, sdk.AccountParameterDataRetentionTimeInDays, "5")
					t.Cleanup(revertParameter)
				},
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithoutDataRetentionSet"),
				ConfigVariables: configVariablesWithoutDatabaseDataRetentionTime(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", r.IntDefaultString),
					checkAccountAndDatabaseDataRetentionTime(t, id, 5, 5),
				),
			},
			{
				PreConfig:       func() { acc.TestClient().Database.UpdateDataRetentionTime(t, id, 20) },
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithoutDataRetentionSet"),
				ConfigVariables: configVariablesWithoutDatabaseDataRetentionTime(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", r.IntDefaultString),
					checkAccountAndDatabaseDataRetentionTime(t, id, 5, 5),
				),
			},
			{
				PreConfig: func() {
					_ = acc.TestClient().Parameter.UpdateAccountParameterTemporarily(t, sdk.AccountParameterDataRetentionTimeInDays, "10")
				},
				ConfigDirectory: acc.ConfigurationDirectory("TestAcc_Database_DefaultDataRetentionTime/WithDataRetentionSet"),
				ConfigVariables: configVariablesWithDatabaseDataRetentionTime(3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_database_old.test", "data_retention_time_in_days", "3"),
					checkAccountAndDatabaseDataRetentionTime(t, id, 10, 3),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func dbConfig(prefix string) string {
	s := `
resource "snowflake_database_old" "db" {
	name = "%s"
	comment = "test comment"
}
`
	return fmt.Sprintf(s, prefix)
}

func dbConfig2(prefix string) string {
	s := `
resource "snowflake_database_old" "db" {
	name = "%s"
	comment = "test comment 2"
	data_retention_time_in_days = 3
}
`
	return fmt.Sprintf(s, prefix)
}

func dbConfigWithReplication(prefix string, secondaryAccountName string) string {
	s := `
resource "snowflake_database_old" "db" {
	name = "%s"
	comment = "test comment 2"
	data_retention_time_in_days = 3
	replication_configuration {
		accounts = [
			"%s"
		]
	}
}
`
	return fmt.Sprintf(s, prefix, secondaryAccountName)
}

// TODO [SNOW-936093]: this is used mostly as check for unsafe execute, not as normal check destroy in other resources. Handle with the helpers cleanup.
func testAccCheckDatabaseExistence(t *testing.T, id sdk.AccountObjectIdentifier, shouldExist bool) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		_, err := acc.TestClient().Database.Show(t, id)
		if shouldExist {
			if err != nil {
				return fmt.Errorf("error while retrieving database %s, err = %w", id, err)
			}
		} else {
			if err == nil {
				return fmt.Errorf("database %v still exists", id)
			}
		}
		return nil
	}
}

func testAccCheckIfDatabaseIsReplicated(t *testing.T, id string) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		replicationDatabases, err := acc.TestClient().Database.ShowAllReplicationDatabases(t)
		if err != nil {
			return err
		}

		var exists bool
		for _, o := range replicationDatabases {
			if o.Name == id {
				exists = true
				break
			}
		}

		if !exists {
			return fmt.Errorf("database %s should be replicated", id)
		}

		return nil
	}
}

func checkAccountAndDatabaseDataRetentionTime(t *testing.T, id sdk.AccountObjectIdentifier, expectedAccountRetentionDays int, expectedDatabaseRetentionsDays int) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		providerContext := acc.TestAccProvider.Meta().(*provider.Context)
		client := providerContext.Client
		ctx := context.Background()

		database, err := acc.TestClient().Database.Show(t, id)
		if err != nil {
			return err
		}

		if database.RetentionTime != expectedDatabaseRetentionsDays {
			return fmt.Errorf("invalid database retention time, expected: %d, got: %d", expectedDatabaseRetentionsDays, database.RetentionTime)
		}

		param, err := client.Parameters.ShowAccountParameter(ctx, sdk.AccountParameterDataRetentionTimeInDays)
		if err != nil {
			return err
		}
		accountRetentionDays, err := strconv.Atoi(param.Value)
		if err != nil {
			return err
		}

		if accountRetentionDays != expectedAccountRetentionDays {
			return fmt.Errorf("invalid account retention time, expected: %d, got: %d", expectedAccountRetentionDays, accountRetentionDays)
		}

		return nil
	}
}
