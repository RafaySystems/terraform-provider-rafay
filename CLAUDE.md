# Rules                                                                                                                                                                   
                                                                                                                                                                          
  ## Nodegroup Sorting Logic                                                                                                                                                
  - State is always stored in alphabetically sorted order (by nodegroup name)                                                                                               
  - Plan/config must be sorted alphabetically before comparison with state                                                                                                  
  - Sorting applies to both `managed_nodegroups` and `node_groups`                                                                                                          
  - Refresh-only should always read from the remote and then store the data in   alphabetically sorted order (by nodegroup name)
  - terraform plan should ensure that only addition/deletion/both/modification either at start/middle/end should display only that is being added/deleted/both/modified.                                                                                                                                                                        

  ## Test Scenarios
  - Scenario 1: User reorders nodegroups in HCL → plan shows no changes
  - Scenario 2: User adds a nodegroup at start/end/middle → only the new one shows as added
  - Scenario 3: User removes a nodegroup at start/end/middle → only the removed one shows as being removed.
  - Scenario 4: User modifies a nodegroup at start/end/middle → only the modified one shows as being modified.


